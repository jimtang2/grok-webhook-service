# Grok-Webhook 

## Description

This service listens to messages sent from [grok-webhook-webext](https://github.com/jimtang2/grok-webhook-webext) to safely replace files within a git repository branch. 

It relies on prompting the LLM to start code blocks with a line in the format: 

```
// webhook$$project_name;file_name;branch_name$$
```

## Setup

1. Define projects name to path map in `config.yml`.
2. Run and configure [grok-webhook-webext](https://github.com/jimtang2/grok-webhook-webext) on browser.
3. Prompt LLM to insert header line in code blocks with project, file and branch.
4. Make sure project head branch matches. If not, files won't be replaced. 

## Example `config.yml`

```yaml
port: :8000
projects:
	project_name_1: /path/to/repo-1/
	project_name_2: /path/to/repo-2/

```

## Example Prompt Header

``` javascript
// webhook$$project_name_1;src/index.js;grok$$

function sayHi() {
	console.log("hi")
}
```

The above prompt response triggers writing to `/path/to/repo-1/src/index.js` on the `grok` branch. Note for security, the HEAD branch must be set to `grok`.
