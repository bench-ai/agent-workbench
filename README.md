![alt text](https://assets-global.website-files.com/65f326ba9a31d0bec3e77363/65f40a01bd69274c25bc035d_Group%2043.png)

# Bench AI Agent Workbench

## About

The agent workbench allows users to create LLM based agents that have access to the internet. We do this
by providing an easy to use programmatic interface for accessing the browser, along with builtin LLM support

## Installation

### Requirements

GO version 1.22.1

### Build from Source

#### Linux / MacOS
Execute the installation script
```shell
# Make Shell Script executable
sudo chmod +x install.sh 
./install.sh
```

Open a new terminal for effects to take place and check installation works
```shell
agent version
```

#### Windows
😂😂😂😂

## Configuring the Agent
The heart of the agent is the configuration file. Here you will dictate the commands for it to invoke along with
some preconfigured settings.

### Browser
#### Settings
```json
// whether or not the browser should be visible. Great for debugging purposes.
{
  "headless": false
}
```

```json
// The max execution time in seconds(optional)
{
  "timeout": 10
}
```

#### Commands
```json
// Opens a webpage in the browser
{
  "command_name": "open_web_page",
  "params": {
    "url": "https://bench-ai.com"
  }
}
```

```json
// Takes a screenshot of the page
// quality: the higher the clearer the image (more comput is needed)
// snap_shot_name: the subfolder name in the resources directory that will contain the saved data
// name: name given to the file in the snapshot folder
{
  "command_name": "full_page_screenshot",
  "params": {
    "quality":90,
    "name": "fullpage.png",
    "snap_shot_name": "s1"
  }
}
```

```json
// Takes a screenshot of a particular element
// scale: how zoomed the image will be
// snap_shot_name: the subfolder name in the resources directory
// name: the subfolder name in the resources directory that will contain the saved data
// selector: xpath to the element
{
  "command_name": "element_screenshot",
  "params": {
    "scale": 2,
    "name": "element.png",
    "selector": "(//img[@class='thumb-image loaded'])[5]",
    "snap_shot_name": "s1"
  }
}
```
```json
// Collects metadata on html elements
// snap_shot_name: the subfolder name in the resources directory that will contain the saved data
// selector: the element of which to extract nodes from
{
  "command_name": "collect_nodes",
  "params": {
    "selector":"body",
    "wait_ready": false,
    "snap_shot_name": "s1"
  }
}
```
```json
// Clicks on a element
// query_type: 
  // search: xpath
// selector: the xpath of the element of which to click on
{
  "command_name": "click",
  "params": {
    "selector":"/html[1]/body[1]/div[1]",
    "query_type": "search"
  }
}
```
```json
// Saves HTML of webpage
// snap_shot_name: the subfolder name in the resources directory that will contain the saved data
{
  "command_name": "save_html", 
  "params": {
    "snap_shot_name": "s1"
  }
}
```
```json
// Sleep for x amount of time
// seconds: how long to sleep for
{
  "command_name": "sleep",
  "params": {
    "seconds": 1
  }
}
```
### LLM 
...

### Example
```json
{
  "operations": [
    {
      "type": "browser",
      "settings" : {
        "timeout": 5,
        "headless": false
      },
      "command_list": [
        {
          "command_name": "open_web_page",
          "params": {
            "url": "https://bench-ai.com"
          }
        },
        {
          "command_name": "full_page_screenshot",
          "params": {
            "quality":90,
            "name": "fullpage.png",
            "snap_shot_name": "s1"
          }
        },
        {
          "command_name": "element_screenshot",
          "params": {
            "scale":2,
            "name": "element.png",
            "selector": "(//img[@class='thumb-image loaded'])[5]",
            "snap_shot_name": "s1"
          }
        },
        {
          "command_name": "collect_nodes",
          "params": {
            "selector":"body",
            "wait_ready": false,
            "snap_shot_name": "s1"
          }
        },
        {
          "command_name": "click",
          "params": {
            "selector":"/html[1]/body[1]/div[1]",
            "query_type": "search"
          }
        },
        {
          "command_name": "save_html",
          "params": {
            "snap_shot_name": "s1"
          }
        },
        {
          "command_name": "sleep",
          "params": {
            "seconds": 1
          }
        }
      ]
    }
  ]
}
```

## CLI
To run the agent simply use the run command and point it to your json file
```shell
agent run ./path/to/my/config.json
```

To execute the agent on a raw json string run
```shell
agent run -j {"operations": [...]}
```


