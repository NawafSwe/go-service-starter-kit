// Package auth provides JWT-based authentication and bcrypt password utilities.
//
// # JWT
//
// Create a ClaimsParser once at startup and reuse it across requests:
//
//	parser, err := auth.NewClaimsParser(cfg.JWT.ISSUER, []byte(cfg.JWT.Secret))
//
// Generate a signed token for a user:
//
//	token, err := parser.GenerateJWTToken(user)
//
// Validate an incoming token and extract the user:
//
//	user, err := parser.ParseJWTToken(rawToken)
//
// Inject / retrieve the authenticated user from a context:
//
//	ctx = auth.SetUserCtx(ctx, user)
//	user := auth.UserFromCtx(ctx)  // returns empty User{} if not present
//
// # Passwords
//
// PasswordGenerator wraps bcrypt — hash on write, compare on login:
//
//	pg := auth.NewPasswordGenerator()
//	hashed, err := pg.EncryptPassword(plaintext)
//	err = pg.ComparePassword(hashed, candidate)
package auth
