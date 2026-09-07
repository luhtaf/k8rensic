# k8rensic

**Digital forensics for Kubernetes, where evidence normally dies with the pod.**

A file is uploaded to an application, the pod is rescheduled, and the file is
gone. Nothing was malicious about the eviction — it is what the platform is for.
But the artefact that would have answered *what happened* no longer exists, and
no amount of log retention brings back bytes nobody stored.

k8rensic captures files where they enter a cluster, preserves them in object
storage, and examines them there.

```
                    ┌──────────────────────────────────────────┐
   HTTP uploads ───►│ corator     application layer            │──┐
                    │ Coraza WAF as a reverse proxy            │  │
                    └──────────────────────────────────────────┘  │
                                                                  ├──►  object
                    ┌──────────────────────────────────────────┐  │     storage
   FTP SMB SMTP ───►│ surisink    network layer                │──┘     (S3)
   SCP  raw HTTP    │ Suricata file extraction on a tap/span   │          │
                    └──────────────────────────────────────────┘          │
                                                                          ▼
                    ┌──────────────────────────────────────────────────────────┐
                    │ s3nitor     examination layer                            │
                    │ IOC · YARA · OTX · VirusTotal · Cuckoo/CAPE detonation   │
                    └──────────────────────────────────────────────────────────┘
                                                                          │
                                                                          ▼
                                                          Elasticsearch · dashboard
```

## Two vantage points, on purpose

The capture components are blind to each other, and that is the design.

`corator` terminates HTTP. It sees exactly what an application receives —
multipart uploads, base64 payloads inside a JSON body, the request path, the
client IP, the headers — because it is the proxy the request passes through. It
cannot see anything that does not go through it.

`surisink` reads a mirror port. It sees files moving over protocols no proxy
terminates: FTP transfers, SMB shares, SMTP attachments, SCP, and HTTP to
workloads nobody thought to put a proxy in front of. It cannot see request
context, because at that layer there is none.

Neither is a superset of the other. Running only one leaves a blind spot that is
easy to describe and easy to forget you have.

## What connects them

Object storage is the seam, and it is deliberately dumb: both capture components
write objects, the examination layer reads them. No component calls another, so
any of the three can be down, restarted, or replaced without the others
noticing.

That is what makes the parts feel like one product rather than a chain that
breaks at the weakest link:

- **A capture component going down loses visibility, not evidence.** Files
  already stored stay stored.
- **The scanner going down loses nothing at all.** It resumes from where it
  stopped — verified by stopping it, uploading, and restarting.
- **New examination is retroactive.** Adding a YARA rule or turning on sandbox
  detonation applies to everything already in the bucket. The bytes are still
  there, which is the entire point of preserving them.
- **A fourth capture source joins by writing objects.** Nothing needs to know it
  exists.

## What each layer does

### corator — application layer

- Transparent reverse proxy: no application changes, no SDK, no redeploy of the
  workload being protected
- Coraza WAF with the OWASP Core Rule Set inspecting every request
- Detects files in `multipart/form-data` **and** base64 payloads embedded in
  request bodies
- Stores to S3-compatible object storage or a local filesystem
- Structured logs to file or Elasticsearch, for SIEM and audit trails
- Clean requests are forwarded to the backend — interception is not blocking

### surisink — network layer

- Tails Suricata's `eve.json` for `fileinfo` events rather than polling the
  filesystem, so it uploads a file when Suricata says it is complete
- Uploads only files Suricata actually stored (`stored=true`)
- Captures FTP, SCP, SMB, SMTP attachments, and HTTP outside any proxied path
- Records SHA-256, MIME type, timestamp, flow id, and source/destination
  endpoints alongside every object
- Deduplicates by SHA-256, in memory or in SQLite for a deployment that restarts
- Passive on a tap, mirror or span port: no path through production traffic
- Configurable worker pool, retry and backoff

### s3nitor — examination layer

- **Event-driven**: a bucket notification triggers the scan, so a file is
  examined seconds after it lands. Falls back to listing a bucket when no broker
  is available.
- **Scanners**: IOC hash lists, YARA rules, AlienVault OTX, VirusTotal, and
  Cuckoo or CAPE sandbox detonation.
- **One document per (file × scanner)**, so a fast scanner never waits behind a
  slow one. A YARA verdict is published in milliseconds while a sandbox is still
  detonating the same file.
- **Synchronous and asynchronous scanners.** A detonation runs for minutes to
  hours in another service: s3nitor submits, keeps the task id, and a separate
  worker collects the verdict later. Nothing holds a queue open for it.
- **Bounded by bytes in flight**, not by file count, so memory is predictable
  regardless of whether the bucket holds a thousand small files or one large
  one.
- **Nothing is scanned twice.** Identity is derived from bucket, key and version,
  so a re-scan happens when content changes and not otherwise.
- **Read-only dashboard** over the findings: counts, severity breakdown,
  filtering, sorting, and coverage over time.
- Publishes to Elasticsearch, Loki, or JSON.

## Getting started

```bash
make build     # compile all three components
make test      # run every test, with -race
make help      # the rest
```

Each component is configured entirely through environment variables and ships as
a container image. See each component's own README under `components/` for its
settings.

## Layout

```
components/
  corator/      application-layer capture
  surisink/     network-layer capture
  s3nitor/      examination and reporting
deploy/         deployment for the whole solution
docs/           architecture, the evidence contract, operations
test/e2e/       tests that cross component boundaries
```

## Roadmap

**Provenance on every finding.** The capture components already record where a
file came from — which sensor, which flow, which HTTP request. Today a finding
says an object matched a rule; it should say *this file arrived from 10.1.2.3 at
14:32 over SMB, and it matched*. For forensics that is the part with legal
weight. The contract is specified in
[docs/evidence-contract.md](docs/evidence-contract.md).

**One install.** A single manifest set that stands up capture, storage,
examination and dashboard together, with the notification wiring already
correct — that wiring is the part that silently does nothing when it is wrong.

**One view.** The dashboard becomes the product surface: not "what did the
scanner find" but "what entered this cluster, how, and what was it".

**Chain of custody.** Object lock and retention at the storage layer, so
preserved evidence is provably unmodified.

## Status

| | |
|---|---|
| Capture at the application layer | working |
| Capture at the network layer | working |
| Event-driven examination, verified end to end | working |
| Sandbox detonation with asynchronous collection | working |
| Findings carry capture provenance | not yet — see roadmap |
| Single deployment for the solution | not yet |

## Licence

MIT, as each component is.
