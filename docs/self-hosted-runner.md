# Overview

This feature of `mapt` allows to setup hosts deployed by it as a GitHub Self Hosted Runner, which can then be directly used for running GitHub actions jobs.
It benefits from all the existing features that `mapt` already provides, allowing to create self-hosted runners that can be used for different QE scenarios.

## Providers and Platforms

Currently, it allows to create self-hosted runners on AWS (Windows Server, RHEL) and Azure (Windows Desktop)

### Prerequisite

To register a Self Hosted Runner for a repository or a GitHub organization, the runner program needs a registration token, which can be obtained by requesting the
GitHub API.

* [Information for requesting a token to register a runner for an Organization](https://docs.github.com/en/rest/actions/self-hosted-runners#create-a-registration-token-for-an-organization)
* [Information for requesting a token to register a runner for a repository](https://docs.github.com/en/rest/actions/self-hosted-runners#create-a-registration-token-for-a-repository)

After obtaining the token we can invoke `mapt` with it to deploy a VM as a Self hosted runner.

For example to add a runner to this repository, we can use the following `curl` command to request a token:

```
% curl -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer <github_personal_access_token>" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/redhat-developer/mapt/actions/runners/registration-token
```
The Response from this `POST` request will be:

```
{
  "token": "ACDZL3QXEIC73UXBDGSEYEI",
  "expires_at": "2024-07-12T19:01:48.478+05:30"
}
```

### Operations

After getting the required token, we need to also decide what we are going to call this runner, the desired name can be passed to the `mapt` command using the
`--ghactions-runner-name` flag.

The full URL of the repository or the GitHub organization also needs to be passed using the `--ghactions-runner-repo` flag.

To deploy a Windows runner on the Azure provider, we can use the following command:

```
% mapt azure windows create --spot \
    --install-ghactions-runner \
    --ghcations-runner-token="ACDZL3QXEIC73UXBDGSEYEI" \
    --ghcations-runner-name "az-win-11" \
    --ghcations-runner-repo "https://github.com/redhat-developer/mapt" \
    --project-name mapt-windows-azure \
    --backed-url file:///Users/tester/workspace \
    --conn-details-output /Users/tester/workspace/conn-details
```
> *NOTE:* additional _labels_ can be added to the runner using the flag `--ghactions-runner-labels`, e.g `--ghactions-runner-lables="azure,mapt,windows"`

