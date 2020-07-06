# run

## Features
- Execute commands wherever you are. Will automatically detect the closest
Runfile.
- Add additional logic to your commands, such as templates and optional
executions.

## Installation
```golang
go get -u -v github.com/jberghoef/run
go install -v github.com/jberghoef/run
```

## Example file
```yaml
echo: echo hello world # Execute a command directly.
home: echo ${HOME} # Use existing environment variables.

test: # Execute a list of commands.
  - echo halt
  - echo and
  - echo catch
  - echo fire

something:
  optional: Would you like to print something? # Mark a command as optional.
  commands:
    - echo say
    - echo {{.Something}} # Use templates to enter values before executing.
    - echo loving

anything:
  command: echo {{.Something}} # Previously entered values will be remembered.

test2:
  env: # Set environment variables before executing.
    ENV_VALUE_2: env2
    ENV_VALUE_3: env3
  commands:
    - echo ${ENV_VALUE_2}
    - echo ${ENV_VALUE_3}
    - :echo # Reference an other command in your file.
```
