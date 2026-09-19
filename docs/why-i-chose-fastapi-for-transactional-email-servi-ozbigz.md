# Why I Chose FastAPI for Transactional Email Service Compliance (Custom Domain Bounces)

A healthtech receipt cannot leave before payment settlement, and its clinical-adjacent data should not leak into an editable provider dashboard. **I chose application-owned templates behind a small Python delivery interface.** The email service owns transport, domain authentication, and bounce events; the application owns the receipt wording, allowed fields, locale, and release history.

TL;DR: Do not pick a transactional email service from an API demo alone. For this job, compare candidates with the same settled-payment fixture and score data location, custom-domain authentication, bounce feedback, idempotent sending, and template change control. Resend, Postmark, and Amazon SES can all be candidates, but the right boundary depends on who must approve copy and where recipient data may be processed.

## What should a transactional email service prove about welcome email compliance?

I first assumed the template editor was a convenience. That framing breaks down when a receipt contains an order identifier, payment total, care location, or support instructions. A dashboard edit can change regulated communication without passing the application repository's review, tests, or deployment record.

That changed my choice.

Keeping the template in the FastAPI codebase gives one pull request a visible diff for copy, variables, and rendering logic. It also makes the provider replaceable at the transport boundary. The trade-off is real: engineers now own preview tooling, localization checks, and HTML compatibility. A nontechnical operations team may reasonably prefer provider-owned templates if it also has an approval workflow and an auditable promotion path. My decision rule is narrow. I keep receipts in the application when engineering and compliance jointly approve changes, fields are derived from a settled order, and a release must reproduce exactly what was sent. I would reconsider when copy changes many times per week and a governed content team, rather than application deploys, is the system of record. The same ownership test applies to a welcome email, although its trigger and allowed data differ from a payment receipt: the team that is accountable for the words needs a controlled way to approve, reproduce, and retire them.

## The focused FastAPI boundary

The payment handler should not render and send inline. Payment systems retry callbacks, networks time out, and a successful send can be followed by a failed acknowledgement. Instead, record settlement and an outbox item in one database transaction. A worker claims that item, renders a versioned template, sends through an adapter, and stores the provider message identifier. A stable idempotency key comes from the order and message purpose, not from a random value created on every retry.

Here is the interface I use in a notebook before wiring any transport. The tiny fake makes the important assertions cheap: no send before settlement, one logical receipt per order, and an explicit template version.

```python
from dataclasses import dataclass
from decimal import Decimal
from typing import Protocol


@dataclass(frozen=True)
class Receipt:
    order_id: str
    recipient: str
    total: Decimal
    currency: str
    template_version: str = "receipt-v3"


class MailTransport(Protocol):
    def send(self, *, message_key: str, to: str, subject: str, html: str) -> str: ...


def deliver_settled_receipt(receipt: Receipt, transport: MailTransport) -> str:
    message_key = f"order:{receipt.order_id}:receipt:{receipt.template_version}"
    subject = "Your payment receipt"
    html = render_receipt(receipt)  # Escapes values and exposes only approved fields.
    return transport.send(
        message_key=message_key,
        to=receipt.recipient,
        subject=subject,
        html=html,
    )
```

The adapter should accept the smallest useful payload. Do not pass the full patient, appointment, or order object because a future template author may expose a field that never belonged in email. This is data minimization enforced by a type boundary rather than a reminder in a wiki.

Three states matter separately: payment settled, delivery accepted by the email service, and final delivery outcome.

Accepted is not delivered.

Store normalized bounce events against the message identifier, verify event authenticity according to the chosen service's documentation, and make event processing idempotent. A permanent bounce should suppress repeated receipts to that address and open an alternate support path; a transient failure can enter a bounded retry policy. The uncomfortable retry case is acceptance followed by a client timeout: the worker has no acknowledgement, yet the service may already have the message. Retrying blindly creates a duplicate receipt. This is why the outbox keeps a stable logical key and why the transport adapter must expose whatever deduplication or reconciliation evidence the selected service supports. The harness should force that timeout, replay the job, and inspect the stored result before the integration is trusted.

## A comparison that starts with evidence

I would run the same evaluation harness against every candidate rather than compare feature-page adjectives. Resend documents domain verification and webhooks. Postmark documents sender signatures, domains, bounce handling, and data-processing terms. Amazon SES documents verified identities, event publishing, and regional endpoints. Those are boundaries to test, not rankings.

| Test | Evidence to capture | Failure that disqualifies a setup |
| --- | --- | --- |
| Domain control | SPF/DKIM verification steps and DMARC alignment result | Production mail cannot align with the organization's From domain |
| Regional review | Contract terms, subprocessors, selected endpoint, and actual event path | The team cannot explain where recipient data and event data are processed |
| Bounce loop | Signed fixture for permanent and transient outcomes | A permanent bounce is retried as if it were temporary |
| Template governance | Diff, approver, immutable version, and rollback exercise | Sent content cannot be reconstructed from an order and version |
| Retry safety | Duplicate settlement event and forced timeout | Two receipts are produced for one logical message |

EU and US compliance is not a checkbox supplied by an email API. The controller or covered organization must determine the applicable role, contract, retention policy, lawful basis, safeguards, and permitted data. GDPR Article 5 includes data minimization and storage limitation. In US health contexts, HIPAA applicability depends on the entities and information involved; the HIPAA Security Rule addresses safeguards for electronic protected health information. Legal and security reviewers must map the actual data flow.

Custom domains deserve a live test too. SPF authorizes sending sources, DKIM signs mail, and DMARC publishes alignment and handling policy. Passing all three does not guarantee inbox placement, but failing alignment is an avoidable configuration defect. I test the exact visible From domain, not a nearby sandbox domain.

## What I measure before copying this choice

The harness begins with 12 cases: settled and unsettled orders; duplicate callbacks; a transport timeout after acceptance; permanent and transient bounces; missing email; Unicode names; two locales; HTML escaping; a stale template version; and an event replay. Twelve is not a magic coverage number. It is a compact set that forces the architecture to reveal where state lives.

Then I track counts and ratios by template version and domain: outbox age, send attempts, accepted messages, permanent bounces, transient bounces, event lag, and duplicate suppression. I do not log rendered bodies or full recipient addresses. Cost belongs in capacity planning, but it is not the primary decision axis; token cost is irrelevant here, while operational review time and retained recipient data are not.

Keep the corpus.

Redacted rendering fixtures, captured MIME structure, and normalized event samples turn provider changes into a regression run instead of a hopeful deployment. Notebook exploration is useful for inspecting a message. Production confidence comes from deterministic fixtures and failure injection.

The concrete choice is therefore conditional: use application-owned, versioned templates for settled-payment healthtech receipts when repository review and reproducibility outweigh dashboard editing speed. Let the delivery service handle transport and feedback, but require evidence for domain alignment, regional data flow, authenticated events, and bounce classification before launch.

## Sources

- https://docs.aws.amazon.com/ses/latest/dg/Welcome.html
- https://resend.com/docs/dashboard/domains/introduction
- https://resend.com/docs/dashboard/webhooks/introduction
- https://postmarkapp.com/developer/user-guide/sender-signatures/sender-signatures-and-domain-verification
- https://postmarkapp.com/developer/webhooks/bounce-webhook
- https://postmarkapp.com/eu-data-privacy
- https://www.rfc-editor.org/rfc/rfc7208
- https://www.rfc-editor.org/rfc/rfc6376
- https://www.rfc-editor.org/rfc/rfc7489
- https://eur-lex.europa.eu/eli/reg/2016/679/oj
- https://www.hhs.gov/hipaa/for-professionals/security/index.html
- https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
