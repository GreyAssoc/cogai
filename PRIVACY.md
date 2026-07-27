# CogAI — Privacy Notice

**Version:** 1.0
**Effective:** 2026-06-09

This Privacy Notice explains what personal data Grey & Associates
Ltd ("we," "us," or "our") processes in connection with the CogAI
product and the online services we operate at getcog.ai. It is
designed to satisfy our obligations under the UK General Data
Protection Regulation (UK GDPR) and the Data Protection Act 2018,
and the EU GDPR where applicable.

If anything in this notice conflicts with the End User Licence
Agreement (`LICENSE`) or the Terms of Service
(`TERMS_OF_SERVICE.md`) as regards personal data, **this Privacy
Notice prevails**.

---

## 1. Who we are (the controller)

  Controller:          Grey & Associates Ltd
  Registered in:       England and Wales
  Company number:      06721769
  Registered office:   Dalton House, 60 Windsor Avenue,
                       London SW19 2RR, United Kingdom
  Contact for privacy: support@getcog.ai

We are the **data controller** for the personal data described
in this notice. We are not currently required to appoint a Data
Protection Officer, but `support@getcog.ai` is the dedicated
contact point for privacy-related correspondence.

---

## 2. The self-hosted CogAI binary processes no data on our behalf

CogAI is a binary you run on hardware you control. Data inside
your CogAI deployment — including prompts, model responses, gear
invocations, persistent facts, and Postgres trace rows —
**never reaches us**. We have no telemetry on CogAI's operation,
no licence-file phone-home, and no remote access to your trace
store.

This means:

- For data that lives inside your self-hosted CogAI deployment,
  **you are the data controller** (and, where your CogAI users
  are other natural persons, you are responsible for your own
  UK GDPR obligations to them).
- We are **not a data processor** for that data and have no
  visibility into it.
- The categories of processing described in the rest of this
  notice cover only the data we process when you interact with
  us directly (e.g. by emailing support, browsing getcog.ai,
  or — in future — purchasing a paid-Tier Subscription).

If you operate CogAI for a hosted deployment we run on your
behalf in a future release, we will act as your processor under
a separate **Data Processing Addendum (DPA)** which we will
publish or provide on request before that service launches.

---

## 3. What personal data we process, and why

### 3.1 Support correspondence

| Item | Detail |
|---|---|
| Categories | Your name (if given), email address, the contents of your email, attachments you choose to send |
| Source | You, when you contact `support@getcog.ai` |
| Purpose | Respond to your enquiry; diagnose problems; keep a record of the interaction |
| Lawful basis | Legitimate interests (UK GDPR Art. 6(1)(f)) — providing support to users of our product. Where you are a paid Subscriber, also Art. 6(1)(b) (performance of contract). |
| Retention | 24 months from the last message in the thread, then deleted, unless we are required to retain it for a longer period under law (e.g. tax records, see §3.4) |
| Recipients / processors | Google Workspace (Google Ireland Ltd / Google LLC) as our email provider |

### 3.2 Website visitors (getcog.ai)

| Item | Detail |
|---|---|
| Categories | IP address, user-agent string, referrer, pages visited, timestamps |
| Source | Your browser, automatically, when you visit the site |
| Purpose | Operate the site; defend it against abuse; understand aggregate usage |
| Lawful basis | Legitimate interests (UK GDPR Art. 6(1)(f)) — running a functional, secure website |
| Retention | Server access logs are retained for 30 days, then deleted or aggregated |
| Recipients / processors | DigitalOcean LLC (hosting), based in the United States; the LON1 region (London) is used where available |
| Cookies | The site uses only strictly-necessary cookies. We do not deploy advertising or third-party analytics cookies. If this changes, a cookie banner will request your consent before any non-essential cookie is set. |

### 3.3 Account and Subscription data (paid Tiers, when launched)

Paid Tiers (Pro, Family, Teams) are not yet generally available.
When they launch we will process the following additional data:

| Item | Detail |
|---|---|
| Categories | Account email, billing name, billing address, country, VAT number (if applicable), Licence-File metadata bound to your account, Subscription status and history |
| Source | You, at sign-up and account-management actions |
| Purpose | Issue and renew your Licence File; provide paid-Tier features; bill you; meet our legal record-keeping duties |
| Lawful basis | Performance of contract (UK GDPR Art. 6(1)(b)) for Subscription operation. Legal obligation (Art. 6(1)(c)) for tax / accounting retention. |
| Retention | Account and Subscription records for the duration of the Subscription, plus 6 years thereafter (UK statute-of-limitations and tax-record requirements) |
| Recipients / processors | Payment processor (to be named in §6 before paid Tiers launch); Google Workspace for billing-related correspondence; our accountants under confidentiality |

We will **not** transmit any data from your self-hosted CogAI
deployment, even after you start paying. The paid Tier unlocks
features in your local binary; it does not introduce telemetry.

### 3.4 Records we are required to keep

Where we are legally required to retain personal data — for
example, invoice records under UK tax law (currently 6 years
from the end of the relevant accounting period) — we will do so
for the period required, then delete or anonymise.

---

## 4. International transfers

We are a UK controller. Some of our processors (e.g. Google,
DigitalOcean) may process personal data outside the UK, including
in the United States.

Where data leaves the UK:

- Transfers to the European Economic Area rely on the European
  Commission's adequacy decision for the UK (and the UK
  Government's adequacy regulations for the EEA).
- Transfers to other third countries (including the US) are
  protected by the UK International Data Transfer Addendum
  (IDTA), or the EU Standard Contractual Clauses with the UK
  Addendum, depending on the processor's contracting model. In
  the case of US processors we also rely, where applicable, on
  the UK Extension to the EU–US Data Privacy Framework.

You can request a copy of the safeguards in place for any given
transfer by emailing `support@getcog.ai`.

---

## 5. Your rights

Under the UK GDPR you have the right to:

- **access** the personal data we hold about you;
- **rectify** inaccurate data;
- **erase** data ("right to be forgotten") in certain
  circumstances;
- **restrict** our processing in certain circumstances;
- **port** data you provided to us in a structured, common
  machine-readable format;
- **object** to processing based on our legitimate interests;
- **withdraw consent** where we relied on consent (currently we
  do not rely on consent for any of the processing above; if we
  do in future, withdrawal will be as easy as giving consent);
- **complain** to the Information Commissioner's Office at
  [ico.org.uk](https://ico.org.uk) if you believe your data has
  been mishandled. You can also complain to a supervisory
  authority in your EU member state of residence where the EU
  GDPR applies.

To exercise any of these rights, email `support@getcog.ai`. We
aim to respond within 5 business days and to complete any
substantive action within one month, as required by the UK GDPR.

We do not charge a fee for handling reasonable requests. We may
refuse or charge for requests that are manifestly unfounded or
excessive, as the law permits.

---

## 6. Processors we use

The processors below act on our written instructions under
contracts that include the appropriate UK GDPR / EU GDPR data-
protection terms.

| Processor | Role | Location |
|---|---|---|
| Google Ireland Limited / Google LLC ("Google Workspace") | Email infrastructure for `support@getcog.ai` | Ireland; United States |
| DigitalOcean, LLC | Hosting of getcog.ai (LON1 region where available) | United States |

A payment processor (e.g. Stripe Payments UK Ltd) will be added
to this list before paid Tiers launch, with the appropriate
processor-level disclosures.

---

## 7. Automated decision-making

We do not use personal data to make decisions about you that
produce legal or similarly significant effects by purely
automated means within the scope of UK GDPR Art. 22.

CogAI itself dispatches your inputs to third-party AI model
providers (which you contract with directly under the BYO-key
model in the EULA). Their automated processing of your inputs is
governed by their own terms and privacy notices, not by this
notice.

---

## 8. Security

We use reasonable technical and organisational measures to
protect personal data, including TLS for data in transit,
access controls on our administrative accounts, and the use of
reputable processors with appropriate certifications (e.g.
ISO 27001 / SOC 2 for the providers listed in §6).

Despite reasonable measures, no system is perfectly secure. If a
personal-data breach affecting you occurs and is likely to
result in a risk to your rights and freedoms, we will notify
you and (where required) the Information Commissioner's Office
in line with UK GDPR Art. 33 / 34.

If you discover a security issue, please report it to
`support@getcog.ai`. We welcome good-faith responsible
disclosure and will not pursue legal action against researchers
acting in good faith within the scope of a reasonable
disclosure.

---

## 9. Children

The Services are not directed at children under 18, except
through the Family-Tier child Seat mechanism. Under that
mechanism, a parent or legal guardian registers Seats on behalf
of children in their household; the parent is responsible for
the child's interaction with CogAI and acts as the relevant
controller for any personal data the child generates in their
self-hosted deployment.

We will delete personal data we discover we have collected from
a child under 18 outside the Family-Tier mechanism without an
appropriate basis.

---

## 10. Changes to this notice

We will publish the current version of this notice at
`https://getcog.ai/privacy` once that page is live, and a copy
will continue to be held in the `PRIVACY.md` file of the public
distribution repository at
[github.com/GreyAssoc/cogai](https://github.com/GreyAssoc/cogai).

Material changes will be notified by email to Account holders
where we have an email address on file, or by an in-Service
notice, at least 14 days before they take effect.

---

## 11. Contact

For any privacy question, to exercise any of the rights in §5,
or to raise a complaint:

  Email:   support@getcog.ai
  Post:    Privacy Officer
           Grey & Associates Ltd
           Dalton House, 60 Windsor Avenue
           London SW19 2RR
           United Kingdom

You may also contact the Information Commissioner's Office:

  Web:     ico.org.uk
  Tel:     0303 123 1113
  Post:    Information Commissioner's Office
           Wycliffe House, Water Lane
           Wilmslow, Cheshire SK9 5AF
           United Kingdom

---

**End of Privacy Notice.**
