# k8rensic

**Digital forensics for Kubernetes, where evidence normally dies with the pod.**

A file is uploaded to an application, the pod is rescheduled, and the file is
gone. Nothing was malicious about the eviction — it is what the platform is for.
But the artefact that would have answered *what happened* no longer exists, and
no amount of log retention brings back bytes nobody stored.

k8rensic captures files at the points where they enter a cluster, preserves them
in object storage, and examines them there. Three runtimes, one chain of
custody.

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

The two capture components are deliberately blind to each other. `corator` sits
in front of an application and sees exactly what that application receives, with
full HTTP context. `surisink` sits on a mirror port and sees everything else —
protocols no proxy terminates, traffic to workloads nobody thought to protect.
Neither is a superset of the other, and running only one leaves a shape of blind
spot that is easy to describe and easy to forget.

## Why one repository

These were three tools that referenced each other in their READMEs and shipped
separately. That is a toolkit, not a product, and the gap shows up in ways that
only appear when you try to use them together:

- **The chain of custody breaks at the seam.** `surisink` records the SHA-256,
  MIME type, flow id, source and destination IP of every file it stores, as S3
  object tags. `s3nitor` never reads them — there is no `HeadObject` or
  `GetObjectTagging` anywhere in its fetcher. So a finding says *this object
  matched a YARA rule* and cannot say *this file was captured from 10.1.2.3 at
  14:32 by the network sensor*. The evidence is in the bucket; the provenance is
  thrown away one component later.
- **A broken component can hide.** `surisink`'s default branch does not build:
  its `go.sum` is missing entries for viper, zap and sqlite, and its
  `PutObjectTagging` call passes a `map[string]string` where minio-go v7 — the
  version it pins itself — requires `*tags.Tags`. Neither repository's CI would
  have noticed, because neither builds the other. Merging them surfaced it in
  one `make build`.
- **Nothing tests the whole path.** All 54 tests in this tree belong to
  `s3nitor`. `corator` and `surisink` have none, and no test anywhere follows a
  file from capture through to a published finding.

So this repository owns the parts that no single component can: the contract
between them, one deployment, and the tests that cross the seams.

## Layout

```
components/
  corator/      application-layer capture   (git subtree of luhtaf/corator)
  surisink/     network-layer capture       (git subtree of luhtaf/surisink)
  s3nitor/      examination and reporting   (git subtree of luhtaf/s3nitor)
deploy/         one deployment for the whole solution
docs/           architecture, the evidence contract, operations
test/e2e/       tests that cross component boundaries
go.work         one workspace over three modules
```

Components are `git subtree`, not submodules and not copies. The code is really
here — `git clone` gets a working tree with no extra step — and history came with
it, so `git log components/s3nitor` still shows why each line is the way it is.
Changes can be pulled from and pushed back to the upstream repositories, which
keep their own releases and their own identity.

Module paths stay `github.com/luhtaf/{corator,surisink,s3nitor}`. Rewriting them
under a k8rensic path would break every existing import and every published
release, and buy nothing a workspace does not already give.

## Getting started

```bash
make build     # compile all three components
make test      # run every test, with -race
make help      # the rest
```

`CGO_ENABLED=1` is set throughout: `s3nitor`'s SQLite driver requires it, and a
build without cgo links a stub that compiles cleanly and panics at startup.

## Where this is going

Ordered by what unblocks the most:

**1. The evidence contract.** One documented shape for a preserved file — object
key convention, the metadata every capture component writes, and the derivation
of the file id — so provenance survives from capture to finding. This is the gap
described above and everything else depends on it. See
[docs/evidence-contract.md](docs/evidence-contract.md).

**2. Provenance in the findings.** `s3nitor` reads the capture metadata and
carries it onto the document, so a finding can answer where the file came from,
which sensor saw it, and over which flow.

**3. One deployment.** Today each component is installed by hand. A single
manifest set stands up capture, storage, examination and dashboard together,
with the notification wiring already correct — that wiring is the part that
silently does nothing when it is wrong.

**4. End-to-end tests.** A file sent through `corator` and a file carried past
`surisink` should both end up as findings. Nothing tests that today, and it is
the only thing that proves the product rather than the parts.

**5. One view.** `s3nitor`'s dashboard already reads findings. Extended with
provenance it becomes the actual product surface: not "what did the scanner
find" but "what entered this cluster, how, and what was it".

## Status

Honest about what is and is not true today:

| | |
|---|---|
| Three components build and test as one tree | yes |
| `s3nitor` runs event-driven against MinIO → Kafka | yes, verified end to end |
| Capture metadata reaches the findings | **no** — item 1 and 2 above |
| One deployment for the solution | **no** — item 3 |
| Any test crossing component boundaries | **no** — item 4 |
| `corator` and `surisink` have tests | **no** |

## Licence

MIT, as each component is.
