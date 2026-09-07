# The evidence contract

The one thing no component can define alone, and the reason this repository
exists.

## The problem, concretely

`surisink` stores a file it extracted from network traffic and tags the object:

```
x-amz-meta-flow_id   2001915
x-amz-meta-src       10.1.2.3
x-amz-meta-dst       10.4.5.6
x-amz-meta-sensor    tap-dmz-01
tags: sha256, mime, ts
```

`s3nitor` then scans that object and publishes:

```json
{ "file_id": "9f2b…", "bucket": "evidence", "key": "2026/09/07/file.1",
  "scanner": "yara", "match": true, "severity": "high" }
```

Everything on the left is missing from the right. `s3nitor` calls `GetObject`
and nothing else — no `HeadObject`, no `GetObjectTagging` — so the provenance is
present in the bucket and absent from the finding. An analyst reading the index
learns that an object matched a rule, and has to go back to the storage layer by
hand to learn where it came from, if the object is still there.

For a forensics product that is the wrong way round. The provenance is the part
with legal weight.

## What has to be agreed

Three things, and only these. Everything else stays each component's business.

### 1. Object key

A key that sorts usefully and never collides:

```
<prefix>/<source>/<yyyy>/<mm>/<dd>/<sha256>[.<ext>]
         │        └── date partition: makes retention and listing cheap
         └── corator | surisink | <sensor id>
```

Content-addressed by SHA-256 rather than named after the original filename. Two
captures of the same bytes converge on one object instead of two, and a filename
chosen by an attacker never becomes a path.

The original filename does not disappear — it moves into metadata, where it is
data rather than structure.

### 2. Capture metadata

Written by whichever component stored the object, on the object itself. Keys are
lowercase and stable; a component that has no value for one omits it rather than
writing an empty string, so "unknown" and "absent" stay distinguishable.

| key | meaning | written by |
|---|---|---|
| `capture_source` | `corator` \| `surisink` | both |
| `capture_sensor` | sensor or pod identity | both |
| `capture_ts` | RFC3339 UTC, when the file was seen | both |
| `sha256` | content hash at capture time | both |
| `mime` | detected content type | both |
| `original_name` | filename as presented | corator |
| `http_host`, `http_path`, `http_method` | request context | corator |
| `client_ip` | request origin | corator |
| `flow_id` | Suricata flow identifier | surisink |
| `src_ip`, `dst_ip`, `src_port`, `dst_port` | the flow | surisink |
| `app_proto` | `http`, `smtp`, `ftp`, `smb`, … | surisink |

Tags versus user metadata is not a free choice: S3 allows **10 tags** per object
and they are mutable and separately readable; user metadata is set at write time
and immutable thereafter. Evidence should be immutable, so the contract puts the
provenance in **user metadata** and leaves tags for operational labels that may
legitimately change.

### 3. File identity

`s3nitor` derives `file_id = sha256(bucket ‖ 0x00 ‖ key ‖ 0x00 ‖ version)` and
that stays as it is — it has to be computable before the object is downloaded,
which rules out the content hash.

But with a content-addressed key, `sha256` is *in* the key, so the two identities
become derivable from each other. That is what lets a question like "has this
exact content been seen before, from any source" be answered without a scan.

## What changes in each component

**corator** and **surisink**: write the metadata keys above. Both already
collect most of it; the work is agreeing on names and moving values from tags to
user metadata.

**s3nitor**: one `HeadObject` on fetch, and carry the metadata onto the
`Finding`. It already fetches the object, so this costs one extra call per
object and no new dependency. The fields land in the document under `capture`,
which the index template must map as `keyword` — the same lesson the `detail`
field already taught, where dynamic mapping produced type conflicts across
scanners.

**k8rensic**: owns this document, and the end-to-end test that proves a file
captured by each component arrives as a finding carrying its provenance.

## Deliberately not in the contract

- **A shared Go package.** Three modules with different dependency pressure —
  Coraza, cgo SQLite, neither — should not be forced to share a type. The
  contract is the wire format: the metadata keys on the object. That is what
  makes it possible for a fourth capture component, in any language, to join by
  writing the right keys.
- **A schema version field.** It would need a migration story before there is
  anything to migrate. Add it when the first incompatible change is actually
  proposed.
- **Signing or WORM.** Both belong in a serious evidence story, and both are
  storage-layer concerns (object lock, retention policies) rather than
  application ones. Worth doing, not worth conflating with this.
