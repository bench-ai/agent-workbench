#!/bin/bash

# Function to add path to file
add_path_to_file_mac() {
    profile_file=$1
    echo "Adding $new_path to $profile_file"
    
    # Check if the path is already in the file
    if grep -q "export PATH=.*$new_path.*" "$profile_file"; then
        echo "Path already exists in $profile_file"
    else
        # Add path to the profile file
        echo "export PATH=\$PATH:$new_path" >> "$profile_file"
        echo "Path added to $profile_file"
    fi
}

#mac installation procedure
install_unix () {

# Define the new path you want to add
new_path="/usr/local/agent/bin"

# Check if the directory exists
if [ ! -d "$new_path" ]; then
    echo "Directory does not exist: $new_path"
    echo "Creating directory..."
    sudo mkdir -p "$new_path"
    
    if [ $? -eq 0 ]; then
        echo "Directory created successfully: $new_path"
    else
        echo "Failed to create directory: $new_path"
        exit 1
    fi
else
    echo "Directory already exists: $new_path"
fi

# build agent
echo "building agent...."
go build .
if [ $? -eq 0 ]; then
    echo " Build Successful" 
else
    echo " Build failed" 
    exit 1
fi

#install 
FILE="agent"
FULLPATH="`pwd`/$FILE"
echo "Installing agent: $FULLPATH in $new_path"
sudo mv $FULLPATH $new_path

# Detect current shell
current_shell=$(echo $SHELL)
echo "$current_shell"

# Add path based on current shell
case $current_shell in
    *"/sh")
        if [ -f "$HOME/.bash_profile" ]; then
            # If .bash_profile exists, use it
            echo " adding to profile bash_profile:$HOME/.bash_profile"
            add_path_to_file_mac "$HOME/.bash_profile"
        elif [ -f "$HOME/.bashrc" ]; then
            # If .bashrc exists, use it
            echo " adding to profile bashrc"            
            add_path_to_file_mac "$HOME/.bashrc"
        else
            # If neither exist, default to .bash_profile
            echo "creating bash_profile"            
            touch "$HOME/.bash_profile"
            add_path_to_file_mac "$HOME/.bash_profile"
        fi
        ;;
    *"/bash")
        if [ -f "~/.profile" ]; then
            # If .bashrc exists, use it
            echo " adding to profile .profile"
            add_path_to_file_mac "~/.profile"
        fi
        ;;
    *"/zsh")
        # Use .zshrc for Zsh
        if [ ! -f "$HOME/.zshrc" ]; then
            touch "$HOME/.zshrc"
            echo "creating zshrc"            

        fi
        add_path_to_file_mac "$HOME/.zshrc"
        ;;
    *)
        echo "Unsupported shell. Please add the path manually."
        ;;
esac

}

# Detect the operating system using `uname` and other methods
OS="$(uname -s)"
case "$OS" in
    Linux*)
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            OS=$NAME
            if [[ "$OS" == "Ubuntu" ]]; then
                echo "This is Ubuntu."
            else
                echo "This is another Linux distribution: $OS."
            fi
        else
            echo "This is a generic Linux system (distro not identified)."
        fi

        export PATH="$PATH:/usr/local/go/bin"

        install_unix
        ;;
    Darwin*)
        echo "This is macOS. Installing in Mac"
        # call install for mac
        install_unix
        ;;
    CYGWIN*|MINGW32*|MSYS*|MINGW*)
        echo "This is Windows."
        ;;
    *)
        echo "Unknown operating system: $OS"
        ;;
esac