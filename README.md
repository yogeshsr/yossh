It’s daunting to remember/find ssh details to the remote machines.
yossh makes it quite easy to setup up aliases for your shell.

## yo ssh
yossh is a utility (golang based) to create aliases to frequently used ssh commands.
yossh reads your infra_file and substitutes user_name tokens defined for the respective env in application.yaml.

See the sample **env.yaml** and **application.yaml**

yossh creates .yo_config with aliases in your home directory and appends (if not already) to your bash_profile & zshrc

## run
Download the [yossh](yossh) file.
    
    # ./yossh /path/to/application.yaml

    or run below if application.yaml file is in the same folder.
    # ./yossh

    Open new terminal to activate aliases.
    To ssh simple run the alias as-
    # p-bla-app-01
    
### Tips
In a team environment it will be useful to have a shared repo for infra_file.
You can create a separate repo (internal/private) to share your env.yaml and update infra_file path in application.yaml

## Develop
If you want to enhance this utility and verify the changes.
    gide install
    go build && go install # or 
    go build -o ./yossh

Pull request are welcome
    

