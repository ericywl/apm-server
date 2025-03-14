---
navigation_title: "Elastic APM"
---

# Elastic APM breaking changes

Before you upgrade, carefully review the Elastic APM breaking changes and take the necessary steps to mitigate any issues.

% To learn how to upgrade, check out <uprade docs>.

% ## Next version

% **Release date:** Month day, year

% ::::{dropdown} Title of breaking change
% Description of the breaking change.
% For more information, check [PR #](PR link).
% **Impact**<br> Impact of the breaking change.
% **Action**<br> Steps for mitigating deprecation impact.
% ::::

## 9.0.0 [9-0-0]

**Release date:** March 25, 2025

::::{dropdown} Change server information endpoint "/" to only accept GET and HEAD requests
This will surface any agent misconfiguration causing data to be sent to `/` instead of the correct endpoint (for example, `/v1/traces` for OTLP/HTTP).
For more information, check [PR #15976](https://github.com/elastic/apm-server/pull/15976).

**Impact**<br>Any methods other than `GET` and `HEAD` to `/` will return HTTP 405 Method Not Allowed.

**Action**<br>Update any existing usage, for example, update `POST /` to `GET /`.
::::

::::{dropdown} The Elasticsearch apm_user role has been removed
The Elasticsearch `apm_user` role has been removed.
For more information, check [PR #14876](https://github.com/elastic/apm-server/pull/14876).

**Impact**<br>If you are relying on the `apm_user` to provide access, users may lose access when upgrading to the next version.

**Action**<br>After this change if you are relying on `apm_user` you will need to specify its permissions manually.
::::

::::{dropdown} The sampling.tail.storage_limit default value changed to 0
The `sampling.tail.storage_limit` default value changed to `0`. While `0` means unlimited local tail-sampling database size, it now enforces a max 80% disk usage on the disk where the data directory is located. Any tail sampling writes that occur after this threshold will be rejected, similar to what happens when tail-sampling database size exceeds a non-0 storage limit. Setting `sampling.tail.storage_limit` to non-0 maintains the existing behavior, which limits the tail-sampling database size to `sampling.tail.storage_limit` and does not have the new disk usage threshold check.
For more information, check [PR #15467](https://github.com/elastic/apm-server/pull/15467) and [PR #15524](https://github.com/elastic/apm-server/pull/15524).

**Impact**<br>If `sampling.tail.storage_limit` is already set to a non-`0` value, tail sampling will maintain the existing behavior.
If you're using the new default, it will automatically scale with a larger disk.

**Action**<br>To continue using the existing behavior, set the `sampling.tail.storage_limit` to a non-`0` value.
To use the new disk usage threshold check, set the `sampling.tail.storage_limit` to `0` (the default value).
::::
