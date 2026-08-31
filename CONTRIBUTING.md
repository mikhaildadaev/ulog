# Contributing to ULOG

🎉 Thank you for being here! We welcome any contribution: from bug reports and feature suggestions to pull requests.

This guide will help you get started.

## Reporting Bugs

1.  **Check for existing issues.** Search for similar issues in the [Issues](https://github.com/mikhaildadaev/ulog/issues) section.
2.  **Create a new issue.** If you don't find one, create a new issue. Please include:
*   What you were doing.
*   What you expected to happen.
*   What actually happened.
*   Your Go version and operating system.
*   A minimal code example to reproduce the problem.

## Suggesting Enhancements

1.  **Discuss first.** For new features or significant changes, please open an issue to discuss your proposal before starting to code. This saves your time if the feature is not a good fit.
2.  **Be specific.** Describe the problem your feature solves and how you propose to solve it.

## Pull Request Process

We use the standard GitHub Flow. Please follow these steps:

1.  **Fork the repository** and create your branch from `main`.
```bash
    git checkout -b feature/amazing-feature
```
2. Write tests. For new features or bug fixes, please add tests. Ensure the test suite passes:
```bash
go test ./...
```
3. Follow the code style. Run go fmt to format your code.
4. Sign your commits. This is required for the Developer Certificate of Origin (DCO). Every commit must be signed:
```bash
git commit -s
```
5. Open a pull request. Open a PR from your branch to the main branch of the main repository.
6. Describe your changes. In the PR description, please include:

What has changed.
A link to the related issue (if any).
Notes for the reviewer.

# Legal Requirements

## Contributor License Agreement (CLA)

**Important:** The ulog projects are dual-licensed (AGPLv3 + Commercial). To protect this model, we require a signed Individual Contributor License Agreement (CLA) for all significant contributions.

Please read the full [CLA](https://github.com/mikhaildadaev/ulog/blob/main/CLA.md)
How to sign: The signature process is described in the CLA file. Typically, this is done by adding a comment to your PR, e.g., "I have read and agree to the CLA", or via a bot (like CLA Assistant).

## Code of Conduct

We strive to maintain an open and welcoming environment. By participating in this project, you agree to abide by the Contributor Covenant Code of Conduct.

## Questions?

If you have any questions, feel free to ask in the Issues or contact the author directly: [mikhaildadaev@mail.ru](mailto:mikhaildadaev@mail.ru)