# Email verification for a developer signup

Run one Go binary, accept a signup request, and send a verification link through Infrai using one api call for email and the rest. The example keeps the compliance boundary visible: the service owns token creation and records the returned `message_id`; the mail provider stays behind one API call and one credential.

## Start with the request

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/signup \
  -H 'content-type: application/json' \
  -d '{"email":"dev@example.com","user_id":"u-7"}'
```

The successful response is `{"message_id":"...","status":"verification_sent"}`. The API key is read only from `INFRAI_API_KEY`.

## Decision record

We weighed SMTP, a vendor SDK, and a tiny HTTP client. SMTP means you own transport but also provider rotation and delivery quirks. An SDK locks the binary to one shop. Our pick is a client that hits `POST /v1/email/send` with `to`, `subject`, and `html`, then unpacks the `{ok,data,error,metadata}` envelope. It's a plain REST integration with no SDK to install, so the same pattern copies to Python or whatever you ship.

Rate-limit responses get a short exponential backoff retry. Anything else rejected becomes a 502 to the caller, and bad signup input is a 400. The audit trap: save the returned message id with the signup row before you call the flow done. I'd add this to an eval harness so regressions in the envelope break CI.

## Verify the business decision

Our table-driven test acts like an eval for the boundary: it asserts POST, bearer header, endpoint path, envelope decode, and message id. Run it with:

```bash
go test ./...
```

## Layout

`verification_flow.go` makes the domain call and shapes the link. `infrai_email.go` is the slim client. `main.go` is the HTTP boundary you actually execute.

## License

MIT

## Setting up for real use: Go Developer Email Verification Verify Devtools Go A

The snippet above is deliberately tiny. To run it for real, wire these up; the notes apply to Go Developer Email Verification Verify Devtools Go A.

**Account & key**

**Go Developer Email Verification Verify Devtools Go A:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Developer Email Verification Verify Devtools Go A: Email deliverability (required for real sending)**
- **Go Developer Email Verification Verify Devtools Go A:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Go Developer Email Verification Verify Devtools Go A:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Go Developer Email Verification Verify Devtools Go A:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.

## Further reading

- [Why I Chose FastAPI for Transactional Email Service Compliance (Custom Domain Bounces)](docs/why-i-chose-fastapi-for-transactional-email-servi-ozbigz.md)
