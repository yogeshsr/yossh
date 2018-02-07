It’s daunting to remember/find ssh details to the remote machines.
It’s quite easy to setup up aliases for your shell.

## yo ssh
yossh is a utility (golang based) to create aliases to frequently used ssh commands.
yossh reads infra_file (application.yaml) to get your infra details.

See the sample **env.yaml**
yossh create .yo_config with aliases in your home directory and appends (if not already) to your bash_profile

You can define user (application.yaml).
//TODO currently it only support one user to be parameterised

## run
    ./yossh

    source ~/.bash_profile # or start open new terminal
    
    # To ssh simple run the alias as
    p-bla-app-01
    
### Tips
In a team environment it will be useful to have a shared repo of aliases.
You can create a separate repo (internal/private) to share your env.yaml and update infra_file(application.yaml)

## Develop
If you want to enhance this utility and verify the changes.
    gide install
    go build && go install # or 
    go build -o ./yossh

Pull request are welcome
    

