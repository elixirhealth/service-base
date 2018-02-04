## Bootstrap a new service repo

Create the repo in GitHub under []elxirhealth org](https://github.com/elxirhealth) and clone locally.

Branch `develop` off `master` branch and push to GitHub.
```bash
git branch -b develop
git push origin develop
```
In repo settings in GitHub, 
- in "Branches", turn on branch protection for both `master` and `develop` and make `develop` default branch
- in "Options", uncheck options for allowing merge commits and rebase merging (leaving just squash merging)

Add new repo to CircleCI [elxirhealth org](https://circleci.com/gh/elxirhealth). Initial build will fail
because there's no config. That's ok. 

In the settings for the new CircleCI project you just created, add a new environment variable `GCR_SVC_JSON`
with a value pasted from
```bash
cat ~/.gcloud/keys/elxir-core-infra.container-registry-ro.json | pbcopy
```
In "Checkout SSH Keys" add an SSH key based on your user, so the project has access to the same repos you do.

From within (basically empty) repo, bootstrap all the goods with
```bash
../service-base/.bootstrap/run.sh MyService
```
where `MyService` is the CamelCase name of your service (often just a single word).

Init deps
```bash
dep init
```

Start fleshing out simple GRPC api and then run
```bash
make proto
```
to make sure things work ok. Then you can run 
```bash
make test
```
to make sure all the tests pass. Once that works, you can push this bootstrapped stuff on a branch
```bash
git checkout -b feature/initial-bootstrap
git add .
git commit -m "initial bootstrap"
git push origin feature/initial-bootstrap
```
You should see the build
