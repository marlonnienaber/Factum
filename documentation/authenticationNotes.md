# Authentication design

## 1. Chosen design

### Model

An account is created only after the user proves control of an allowlisted email address. The password is chosen after that proof. Sign-up creates no session; only sign-in does.

Before sign-up is completed, the server stores an email verification token record, not a user:

```text
email                   unique, canonical email address
token_hash              unique SHA-256 hash of the current token
expires_at              token expiration time
minutely_window_started_at start of the current one-minute window
hourly_send_count       accepted sends in the current one-hour window
hourly_window_started_at start of the current one-hour window
daily_send_count        accepted sends in the current 24-hour window
daily_window_started_at start of the current 24-hour window
```

The plaintext token consists of 32 cryptographically random bytes encoded as 64 lowercase hexadecimal characters. It is sent by email and never stored. Each token expires after 15 minutes. There is at most one token record per email. The initial email and every accepted resend count toward the minutely, hourly, and daily send limits.

The email link is `{APP_URL}/sign-up?token={token}`. `APP_URL` is trusted configuration rather than a value derived from the request. In development it defaults to `http://localhost:{CLIENT_PORT}`, with `CLIENT_PORT` defaulting to `5173`. The server listens on the separately configured `SERVER_PORT`, which defaults to `3001`.

The design assumes that the intended user controls the mailbox and does not disclose the link. It also assumes HTTPS, a trusted application and email provider, a fixed configured origin for email links, and consistent email canonicalization across the allowlist, token records, users, and sign-in.

### Scenario A: typical sign-up with one resend

| Step | User experience | System behavior |
| --- | --- | --- |
| 1 | The user enters an email on the public sign-up page. | The client calls `POST /api/verify-email`. |
| 2 | After a successful verification request, the user reaches the same “check your email” view. Request failures show an error toast and leave email entry available. | The server canonicalizes the email. If it is allowlisted, has no user, and is eligible for issuance, the server creates a random token, stores its hash and 15-minute expiry, and sends the link. Every syntactically valid request receives the same response. |
| 3 | After the local 60-second countdown, the user selects resend and sees the same neutral confirmation. | If another send is allowed, the server replaces the token hash, resets the 15-minute expiry, updates the send limits, and sends a new link. The previous link becomes invalid. |
| 4 | The user opens the newest email link and sees the choose-password page. | The presence of a token query parameter selects the password step. Opening the link does not validate or consume the token, create a user, or create a session. |
| 5 | The user chooses and confirms a password. | The client validates the token’s 64-character hexadecimal format, password length, and password confirmation before sending the token and password to `POST /api/sign-up`. Invalid token format shows an error toast; password errors appear on the relevant inputs. The server checks token existence, expiry, and the current allowlist; the unique email constraint prevents duplicate accounts. In one transaction, it consumes the matching unexpired token record and inserts a verified user with an Argon2id password hash. The endpoint returns `201 Created` without a response body or session. |
| 6 | The user sees “Sign-up complete. Sign in to continue.” and is redirected to sign-in. | Sign-up has not created or changed a session. |
| 7 | The user enters the new credentials and enters the application. | `sign-in` verifies the password and creates the session. |

The sign-up route sets `<meta name="referrer" content="no-referrer" />` and does not persist the token in client storage. The shared layout still loads external fonts. The development email service logs verification links instead of sending email. Production token-log redaction and email delivery are deployment requirements, not guarantees of the current implementation.

The check-email view displays the entered email without client-side normalization:

> Check the inbox of {email}. Open the link in the latest mail to complete the sign up. Check spam too.

A successful resend request shows “Request received. Check your inbox for the latest link.” Neither the view nor the response confirms actual delivery. A back arrow returns to a fresh email-entry step.

The client uses one `/sign-up` route with separate email-entry, check-email, and password components. The parent selects them through a writable derived `SignUpStep` union. Each component owns its inputs, validation, or resend state; markup and styles are defined explicitly in each component. All three cards display “Sign up” as their heading.

### Scenario B: incomplete sign-up and restart

| Step | User experience | System behavior |
| --- | --- | --- |
| 1 | The user requests a link but leaves without completing sign-up. | The token expires after 15 minutes. Its record is retained while its send-limit state is still needed. No user or session exists. |
| 2 | The user later submits the email again, either from sign-up or the resend page. | Both actions call `POST /api/verify-email` and have identical semantics. |
| 3a | If another send is allowed, the user receives a new link. | The server rotates the token, resets its expiry, updates the send limits, and invalidates the previous link. |
| 3b | If sending is limited, the user still sees the same neutral response but receives no email. | The server leaves the record and current token unchanged. Token expiry does not bypass the send limits. |
| 4 | The user completes sign-up through a valid current link. | The normal sign-up transaction consumes the record and creates the user. |

For each email, the server permits at most one email per one-minute window, five emails per one-hour window, and ten emails per 24-hour window. Each window independently begins with the first accepted send after its previous window expires. Each stores a window start. The one-minute window allows only one send, so an active window is already full and needs no counter; every accepted send sets its start to the current time. The hourly and daily windows also store send counts. An accepted send increments the count of an active hourly or daily window or resets an expired window to a count of one and a start of the current time. Sending is allowed only when all three windows permit it. The initial email and accepted resends use the same limits. A rejected request does not rotate the token, send email, or alter any window start or hourly or daily send counter.

An expired record is retained while any minutely, hourly, or daily send-limit state is still needed. After all three windows expire, it may be deleted. Retaining the record prevents token expiry from resetting the limits.

The server never returns whether an email was sent or when another send becomes available. The check-email component initializes its local windows when it mounts after the initial verification request succeeds, counting that request as one. Its first 60-second countdown therefore starts after the response arrives. Subsequent resend attempts update the minute start and hourly and daily counts when triggered, even if the request fails. It permits at most five local requests in one hour and ten in 24 hours. The email-entry component does not track local windows for failed initial requests. When a local limit is reached, it prevents another request until that local window expires and explains the public policy without claiming that an email was sent.

Going back unmounts check-email and clears its local windows; submitting an email again starts fresh, even for the same address. Refreshing the page, leaving the route, or using another browser also clears this state. This does not bypass security because the server independently enforces the per-email limits. The client behavior is advisory and does not reflect allowlist membership, user existence, token-record existence, actual sends, or server throttling state.

Because tokens expire after 15 minutes, a user who reaches a server limit may need to wait for a new send window after the current token expires.

### Scenario C: failed sign-in followed by a sign-up restart

| Step | User experience | System behavior |
| --- | --- | --- |
| 1 | A user with an incomplete sign-up attempts to sign in. | `sign-in` checks only `users`. A token record is not an account and contains no password, so sign-in fails like any nonexistent-user attempt. |
| 2 | The user sees an error toast: “Email or password incorrect.” | The endpoint returns `invalid_credentials` without revealing whether a user or token record exists. |
| 3 | The user returns to sign-up and submits the email again. | `POST /api/verify-email` rotates the existing token if another send is allowed; otherwise it returns the same neutral response without changing state. |
| 4 | The user follows a valid link, chooses a password, and then signs in. | Sign-up and sign-in proceed as in Scenario A. |

### Scenario D: opening an invalidated confirmation link

| Step | User experience | System behavior |
| --- | --- | --- |
| 1 | The user requests a resend and later opens an older email. | An accepted resend has replaced the stored token hash, so the older token no longer matches the record. |
| 2 | The user submits the password and sees an error toast; the password view and entered values remain available. | Token format is validated locally on submission. A well-formed unknown, rotated, expired, or consumed token is rejected by sign-up with the same generic response. The client does not redirect or present a special restart view. |
| 3 | The user can return to `/sign-up` without a token to request another link. | The normal per-email send limits apply. The response remains neutral whether a new token is sent or not. |

### Scenario E: repeated resend requests from the client

| Step | User experience | System behavior |
| --- | --- | --- |
| 1 | After requesting an email, the user sees a 60-second countdown before resend is available. | The first countdown starts when check-email mounts; subsequent countdowns start when resend is triggered. Dates and duration arithmetic use `@internationalized/date` in UTC, with a daily duration of exactly 24 hours. The countdown is based only on local state. The server independently enforces its own one-minute window with a maximum of one accepted send. |
| 2 | The user makes five requests within the local one-hour window. | The client prevents more requests until that local window expires. It derives this behavior only from its local timestamps and receives no server state. |
| 3 | The user reaches ten requests within the local 24-hour window. | The client prevents more requests until that local window expires. The server independently permits at most ten accepted sends in its own 24-hour window. |
| 4 | The result looks the same whether a user already exists for the email or no user exists. | For an existing user, the server sends nothing. For an eligible email without a user, the server rotates and sends only if its independent limits allow it. Neither outcome is disclosed. |
| 5 | Going back, leaving or refreshing the page, or using another browser removes the client-side restriction. | This does not bypass security. The server still enforces the per-email limits and returns the same neutral response. |

### Endpoint outcomes

| Endpoint | Input errors | Business-state outcomes | Success |
| --- | --- | --- | --- |
| `POST /api/verify-email` | Malformed JSON or an invalid email format produces the normal validation response. | Every syntactically valid request receives `204 No Content`. It does not reveal allowlist membership, user or token-record existence, token rotation, throttling, email delivery, or remaining delay. | The client opens the same “check your email” page. |
| `sign-up` | Malformed JSON, an invalid password, or a token that is not exactly 64 hexadecimal characters produces the normal validation response. | An unknown, expired, rotated, consumed, no-longer-allowlisted, already-used, or otherwise unusable well-formed token produces `400 Bad Request` with `sign_up_not_available`. | The token record is consumed, a verified user is created, and the endpoint returns `201 Created` without a response body or session. The client goes to sign-in. |
| `sign-in` | Malformed input produces the normal validation response. | A missing user, token-record-only email, wrong password, or unavailable account all produce `invalid_credentials` with the same status, response shape, and approximate timing. A fixed dummy Argon2 hash is checked when no user exists. | The server creates a session. |

Unexpected database failures return a generic server error without implementation details. Email-delivery failures are logged internally, but the endpoint still returns `204 No Content` so delivery attempts cannot reveal whether an email was eligible.

### Traffic protection

The application enforces the minutely, hourly, and daily per-email window limits because it must canonicalize the email and update the token record transactionally.

The reverse proxy handles general traffic protection: per-IP and global request limits, connection limits, body and header limits, timeouts, and HTTPS. It must redact verification tokens and full verification URLs from logs. Forwarded client addresses are trusted only from the configured proxy.

## 2. Why the design works this way

### Prove email ownership before choosing a password

If credentials are selected first, an attacker can start sign-up for a victim with an attacker-known password and hope the victim clicks the legitimate verification email. Choosing the password after opening the email link prevents the attacker from selecting or influencing the credentials. It also keeps password hashes out of token records and URLs.

### Store a token record, not an unverified user

An unverified user would reserve the unique email and let anyone who knows an allowlisted address block its owner. A token record grants no application access and does not reserve the email in `users`. It also removes the authenticated-but-unverified state from the server and client.

### Use a stateful opaque token

Consuming the token record makes replay protection independent of the resulting user. If that user is later deleted, the consumed token cannot recreate the account. State also permits immediate rotation and revocation.

The finite exact-email allowlist and unique `email` constraint bound the number of token records. This makes the storage exposure acceptable for this application.

### Store only the token hash

A database reader cannot use the stored SHA-256 hash as the bearer token. Argon2 is unnecessary for a random 256-bit token: unlike a human password, it is already infeasible to guess even when hash checks are fast.

### Rotate under standard per-email send limits

Rotation gives a user a fresh link and invalidates an older one. A one-minute window allowing one email prevents rapid rotation. One-hour windows allowing five emails and 24-hour windows allowing ten emails bound sustained database writes, token invalidation, and email abuse. Rejected requests preserve the existing link and cannot be used solely to invalidate it. Keeping the true server state out of responses avoids leaking allowlist membership and recent request activity.

### Do not consume the token on `GET`

Email security systems may open links automatically. Consumption occurs only when the password is submitted, so a scanner cannot create an account or invalidate the link.

### Consume and create atomically

Token deletion and user insertion occur in one database transaction. Only one concurrent submission can consume the record. If insertion fails, rollback restores it. The unique constraint on `users.email` remains the final concurrency safeguard.

### Create sessions only through sign-in

The verification token authorizes account creation, not authentication. Only `sign-in` creates sessions, leaving one session-creation boundary to secure, test, and reason about. The additional step also verifies the newly chosen credentials before access is granted.

### Use generic authentication responses

Uniform responses prevent the public endpoints from revealing allowlist membership, account existence, incomplete sign-up, or token state. Explicit validation errors remain safe because they describe only the caller's submitted format.

## 3. Alternatives not selected

### Create an unverified user before email verification

This lets an attacker reserve another person's allowlisted email and requires an authenticated-but-unverified state. The chosen design stores only a non-authenticating token record until ownership is proven.

### Put the email or password hash in a stateless token

A stateless token removes provisional database state, but its replay protection depends on the created user continuing to exist until token expiry. Deleting that user could make the token usable again. Older stateless tokens also cannot be consumed or invalidated when a newer token is issued. Without per-email server state, the application also cannot track how many emails it has sent to an address, so the endpoint can be weaponized for repeated email spam. Adding that tracking elsewhere would make the flow stateful again. Encryption adds key and nonce management. The bounded allowlist makes stateful records the lower-risk choice here.

### Choose the password before email verification

This permits attacker-selected credentials before ownership is proven and requires the password hash to be stored or transported during the pending phase. Choosing the password afterward is both simpler and resistant to that pre-hijacking pattern.

### Require the original password again after opening the link

Requiring both the link and a previously chosen password would also prevent pre-hijacking, but adds password re-entry and retains a password hash during the pending phase. Letting the verified mailbox owner choose the password provides the better usability trade-off under the stated email-ownership assumption.

### Sign the user in when sign-up completes

Automatic sign-in would make the email token part of session establishment and add another session-creation path with additional CSRF, session-fixation, account-switching, and failure-handling considerations. The UX benefit does not justify that complexity. Redirecting to sign-in keeps sign-up and authentication separate.

### Return the server's exact send-limit state

An exact delay or remaining allowance can reveal whether a user has completed sign-up for a given email: requesting a link creates a token record only when no user exists, so returned limit state would expose that difference. It could also reveal recent token activity. The chosen client-only countdown and limits explain and enforce the public policy locally without exposing actual server state.
