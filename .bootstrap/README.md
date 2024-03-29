## Bootstrap a new service repo

Create the repo in GitHub under [elixirhealth org](https://github.com/elixirhealth) and clone locally.

Branch `develop` off `master` branch and push to GitHub.
```bash
git checkout -b develop
git push origin develop
```
In repo settings in GitHub, 
- in "Branches", turn on branch protection for both `master` and `develop` and make `develop` default branch
- in "Options", uncheck options for allowing merge commits and rebase merging (leaving just squash merging)

Add new repo to CircleCI [elixirhealth org](https://circleci.com/gh/elixirhealth). Initial build will fail
because there's no config. That's ok. 

In the settings for the new CircleCI project you just created, import the environment variable `GCR_SVC_JSON`
from that in the `courier` project.

In "Checkout SSH Keys" add an SSH key based on your user, so the project has access to the same repos you do.

From within (basically empty) repo, bootstrap all the goods with
```bash
../service-base/.bootstrap/run.sh MyService
```
where `MyService` is the CamelCase name of your service (often just a single word).

Get the dependencies, install the git hooks, generate code for the (empty) grpc API, and auto clean 
up code:  
```bash
make get-deps install-git-hooks proto fix
```
Confirm tests pass and code lints ok 
```bash
make test lint
```
Once those work, you can push this bootstrapped stuff on a branch
```bash
git checkout -b feature/initial-bootstrap
git add .
git commit -m "initial bootstrap"
git push origin feature/initial-bootstrap
```
You should see the build.

If using Goland as your IDE, in Preferences -> Go -> Imports, set "Sorting type" to "goimports" and 
check all the boxes. in Preferences -> Tools -> File Watchers, add a "goimports" file watcher (with 
default settings).

