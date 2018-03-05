# Command-line wrapper for AWS Batch

As described in the
[AWS Batch At Fred Hutch documentation](https://fredhutch.github.io/aws-batch-at-hutch-docs/),
Fred Hutch users must use a wrapper for some aspects of interacting
with AWS Batch (submitting, terminating, and canceling jobs).

This wrapper is meant for command-line use only.
The [Python wrapper](https://github.com/FredHutch/fredhutch_batch_wrapper) is in
a separate repository.


## Installation

Note that this wrapper is **already installed** on the `rhino` and `gizmo` systems.
The command `batchwrapper` is already in your `PATH`.

For other systems, follow the instructions on the
[releases page](https://github.com/FredHutch/batch-wrapper-client-golang/releases).

## Prerequisites

You must have previously
[obtained your AWS credentials](https://teams.fhcrc.org/sites/citwiki/SciComp/Pages/Getting%20AWS%20Credentials.aspx)
and [requested access](https://fredhutch.github.io/aws-batch-at-hutch-docs/#request-access)
to AWS Batch.

## Usage

Usage is similar to that of the [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/batch/index.html) for Batch.

In fact, you should continue to use the CLI for everything except
submitting jobs (`submit-job`), terminating jobs (`terminate-job`),
and canceling jobs (`cancel-job`).

Running the `batchwrapper` command without any arguments gives brief help:

```
usage: batchwrapper [<flags>] <command> [<args> ...]

Cancel, terminate, and submit AWS Batch jobs. Full docs at
https;//bit.ly/HutchBatchDocs

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  cancel --job-id=JOB-ID --reason=REASON
    Cancel a job you submitted

  terminate --job-id=JOB-ID --reason=REASON
    Terminate a job you submitted

  submit --cli-input-json=JSON_FILE
    Submit a job
```

#### Submitting a job

You can get some help with `batchwrapper submit --help`:

```
usage: batchwrapper submit --cli-input-json=JSON_FILE

Submit a job

Flags:
  --help                      Show context-sensitive help (also try --help-long
                              and --help-man).
  --cli-input-json=JSON_FILE  JSON file containing job info
```

To submit a job, you need to put the pertinent information in a JSON file.
Unlike with the AWS CLI, you don't put the `file://` prefix in front of the
JSON file.

You can prepare a JSON file by following the
[example](https://fredhutch.github.io/aws-batch-at-hutch-docs/#submit-your-job) in the Fred Hutch batch documentation.

Assuming your JSON file is called `job.json`, you can submit it
as follows:

```
batchwrapper submit --cli-input-json job.json
```

The batch wrapper will return the Job ID and name.

#### Terminating a job

Brief help is available with `batchwrapper terminate --help`:

```
usage: batchwrapper terminate --job-id=JOB-ID --reason=REASON

Terminate a job you submitted

Flags:
  --help           Show context-sensitive help (also try --help-long and
                   --help-man).
  --job-id=JOB-ID  Job ID
  --reason=REASON  reason for termination
```

So, for example, you can terminate a job you own with the following
command (replace the job ID with your own):


```
batchwrapper terminate --job-id 13732097-3f5d-42bc-b60f-2cd166486074 \
   --reason "there is a problem"
```

#### Canceling a job

Brief help is available with `batchwrapper cancel --help`:

```
usage: batchwrapper cancel --job-id=JOB-ID --reason=REASON

Cancel a job you submitted

Flags:
  --help           Show context-sensitive help (also try --help-long and
                   --help-man).
  --job-id=JOB-ID  Job ID
  --reason=REASON  reason for termination
```

So, for example, you can cancel a job you own with the following
command (replace the job ID with your own):



```
batchwrapper cancel --job-id 13732097-3f5d-42bc-b60f-2cd166486074 \
   --reason "there is a problem"
```
