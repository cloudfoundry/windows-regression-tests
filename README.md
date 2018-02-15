# Windows Regression Tests

This suite contains regression tests for Windows specific CF features.

## Test Setup
### Prerequisites for running WRTs
- Install golang >= `1.7`. Set up your golang development environment, per
  [golang.org](http://golang.org/doc/install).
- Install the [`cf CLI`](https://github.com/cloudfoundry/cli).
  Make sure that it is accessible in your `$PATH`.
- Install [curl](http://curl.haxx.se/)
- Check out a copy of `windows-regression-tests`
  and make sure that it is added to your `$GOPATH`.
  The recommended way to do this is to run:

  ```bash
  go get -d code.cloudfoundry.org/windows-regression-tests
  ```

  You will receive a warning:
  `no buildable Go source files`.
  This can be ignored, as there is only test code in the package.
- Install a running Cloud Foundry deployment
  to run these acceptance tests against.

## Test Configuration
You must set the environment variable `$CONFIG`
which points to a JSON file
that contains several pieces of data
that will be used to configure the tests.

The following can be pasted into a terminal
and will set up a sufficient `$CONFIG`
to run the core test suites.

```bash
cat > integration_config.json <<EOF
{
  "api": "api.env.cf-app.com",
  "admin_user": "admin",
  "admin_password": "password",
  "apps_domain": "env.cf-app.com",
  "num_windows_cells": 1,
  "skip_ssl_validation": true,
  "isolation_segment_name": "",
  "stack": "windows2012R2"
}
EOF
export CONFIG=$PWD/integration_config.json
```

## Test Execution
To execute the tests, run the following from the root directory of `windows-regression-tests`:
```bash
./bin/test
```
