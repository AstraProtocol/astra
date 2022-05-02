<!-- < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < < ☺
v                               ✰  Thanks for creating a PR! ✰    
v    Before smashing the submit button please review the checkboxes.
v    If a checkbox is n/a - please still include it but + a little note why
☺ > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > > >  -->

## Description

<!-- Add a description of the changes that this PR introduces and the files that
are the most critical to review.
-->

Closes: #XXX

## PR Checklist:

- [ ] Have you read the [CONTRIBUTING.md](https://github.com/AstraProtocol/astra/blob/master/CONTRIBUTING.md)?
- [ ] Does your PR follow the [C4 patch requirements](https://rfc.zeromq.org/spec:42/C4/#23-patch-requirements)?
- [ ] Have you rebased your work on top of the latest master?
- [ ] Have you checked your code compiles? (`make`)
- [ ] Have you included tests for any non-trivial functionality?
- [ ] Have you checked your code passes the unit tests? (`make test`)
- [ ] Have you checked your code formatting is correct? (`go fmt`)
- [ ] Have you checked your basic code style is fine? (`golangci-lint run`)
- [ ] If you added any dependencies, have you checked they do not contain any known vulnerabilities? (`go list -json -m all | nancy sleuth`)
- [ ] If your changes affect the client infrastructure, have you run the integration test?
- [ ] If your changes affect public APIs, does your PR follow the [C4 evolution of public contracts](https://rfc.zeromq.org/spec:42/C4/#26-evolution-of-public-contracts)?
- [ ] If your code changes public APIs, have you incremented the crate version numbers and documented your changes in the [CHANGELOG.md](https://github.com/AstraProtocol/astra/blob/master/CHANGELOG.md)?


______

## Reviewers Checklist

**All** items are required. Please add a note if the item is not applicable and please add your handle next to the items reviewed if you only reviewed selected items.

I have...

- [ ] confirmed the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
- [ ] confirmed all author checklist items have been addressed
- [ ] confirmed that this PR does not change production code

