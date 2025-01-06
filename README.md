# HELM Version Updater Action

A GitHub Action that updates version values in Helm Version files using GitHub's API. This action can modify any specified key in a YAML file within a given repository branch and automatically commit the changes.

## Features

- Updates version values in YAML files without cloning the repository
- Uses GitHub's API for efficient file modifications
- Supports custom branch, file path, and version key specifications
- Provides detailed error messages for troubleshooting

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `repository` | The GitHub repository path | Yes | N/A |
| `github_token` | GitHub token with repository write permissions | Yes | N/A |
| `branch` | The branch to modify | No | main |
| `values_file` | Path to the YAML file within the repository | No | app/values/values.beta.yaml |
| `version_key` | The key within the YAML file to update | No | version |
| `commit_message` | The commit message for the changes | No | update version |
| `version` | The new version value to set | Yes | N/A |

## Usage

```yaml
name: Update Helm Version
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'New version value'
        required: true

jobs:
  update-version:
    runs-on: ubuntu-latest
    steps:
      - name: Update YAML Version
        uses: your-username/yaml-version-updater@v1
        with:
          repository: /your-org/your-repo
          github_token: ${{ secrets.GITHUB_TOKEN }}
          version: ${{ github.event.inputs.version }}
          branch: main
          values_file: app/values/values.beta.yaml
          version_key: version
```

## Testing Locally

1. Build the Docker image:
   ```bash
   docker build -t updater .
   ```

2. Create a test YAML file and set environment variables:
   ```bash
   # Create a test values.yaml
   echo "version: v0.0.1" > values.yaml
   
   # Set required environment variables
   export INPUT_REPOSITORY="/your-org/your-repo"
   export INPUT_GITHUB_TOKEN="your-github-token"
   export INPUT_VERSION="v1.0.0"
   export INPUT_BRANCH="main"
   export INPUT_VALUES_FILE="values.yaml"
   export INPUT_VERSION_KEY="version"
   export INPUT_COMMIT_MESSAGE="update version"
   ```

3. Run the action locally:
   ```bash
   docker run --rm \
     -e INPUT_REPOSITORY \
     -e INPUT_GITHUB_TOKEN \
     -e INPUT_BRANCH \
     -e INPUT_VALUES_FILE \
     -e INPUT_VERSION_KEY \
     -e INPUT_COMMIT_MESSAGE \
     updater
   ```

## Error Handling

The action provides detailed error messages for common issues:

- Invalid repository path
- Missing or invalid GitHub token
- File not found
- Version key not found in YAML
- API request failures

Error messages are formatted as GitHub Action error annotations and will appear in your workflow logs.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request
