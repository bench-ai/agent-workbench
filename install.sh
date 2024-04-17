os=$(uname -s)

variable_name=""

case $os in
    Linux*)
      echo "Linux"
      echo $variable_name
    
      rm -rf "/usr/local/agent"
      mkdir -p "/usr/local/agent/bin"

      export PATH="$PATH:/usr/local/go/bin"

      go build .
      mv agent "/usr/local/agent/bin"
      ;;
    Darwin*)    echo "macOS detected";;
    *)          echo "Unsupported operating system: $os"; exit 1;;
esac