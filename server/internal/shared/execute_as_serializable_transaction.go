package shared

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const maxSerializableTransactionAttemptCount = 3

func ExecuteAsSerializableTransaction[ReturnType any](
	db *sqlx.DB,
	codeToExecute func(transaction *sqlx.Tx) (ReturnType, error),
) (ReturnType, error) {
	var zeroReturnValue ReturnType

	for attempt := 1; attempt <= maxSerializableTransactionAttemptCount; attempt++ {
		transaction, beginTransactionError := db.BeginTxx(
			context.Background(),
			&sql.TxOptions{Isolation: sql.LevelSerializable},
		)
		if beginTransactionError != nil {
			return zeroReturnValue, fmt.Errorf("beginning serializable transaction failed: %w", beginTransactionError)
		}

		returnValue, transactionError := codeToExecute(transaction)
		if transactionError != nil {
			rollbackError := transaction.Rollback()
			if rollbackError != nil && !errors.Is(rollbackError, sql.ErrTxDone) {
				return zeroReturnValue, fmt.Errorf("transaction execution failed: %w; rollback failed: %v", transactionError, rollbackError)
			}
			if isSerializationFailureError(transactionError) {
				continue
			}
			return zeroReturnValue, transactionError
		}

		commitError := transaction.Commit()
		if commitError == nil {
			return returnValue, nil
		}
		if isSerializationFailureError(commitError) {
			continue
		}
		return zeroReturnValue, fmt.Errorf("committing serializable transaction failed: %w", commitError)
	}

	return zeroReturnValue, fmt.Errorf("serializable transaction failed after %d attempts", maxSerializableTransactionAttemptCount)
}

func ExecuteAsSerializableTransactionWithoutResult(
	db *sqlx.DB,
	codeToExecute func(transaction *sqlx.Tx) error,
) error {
	for attempt := 1; attempt <= maxSerializableTransactionAttemptCount; attempt++ {
		transaction, beginTransactionError := db.BeginTxx(
			context.Background(),
			&sql.TxOptions{Isolation: sql.LevelSerializable},
		)
		if beginTransactionError != nil {
			return fmt.Errorf("beginning serializable transaction failed: %w", beginTransactionError)
		}

		transactionError := codeToExecute(transaction)
		if transactionError != nil {
			rollbackError := transaction.Rollback()
			if rollbackError != nil && !errors.Is(rollbackError, sql.ErrTxDone) {
				return fmt.Errorf("transaction execution failed: %w; rollback failed: %v", transactionError, rollbackError)
			}
			if isSerializationFailureError(transactionError) {
				continue
			}
			return transactionError
		}

		commitError := transaction.Commit()
		if commitError == nil {
			return nil
		}
		if isSerializationFailureError(commitError) {
			continue
		}
		return fmt.Errorf("committing serializable transaction failed: %w", commitError)
	}

	return fmt.Errorf("serializable transaction failed after %d attempts", maxSerializableTransactionAttemptCount)
}

func isSerializationFailureError(transactionError error) bool {
	var postgresError *pq.Error
	return errors.As(transactionError, &postgresError) && postgresError.Code == "40001"
}
