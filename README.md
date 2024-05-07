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
// snapshot_name: the subfolder name in the resources directory that will contain the saved data
// name: name given to the file in the snapshot folder
{
  "command_name": "full_page_screenshot",
  "params": {
    "quality":90,
    "name": "fullpage.png",
    "snapshot_name": "s1"
  }
}
```

```json
// Takes a screenshot of a particular element
// scale: how zoomed the image will be
// snapshot_name: the subfolder name in the resources directory
// name: the subfolder name in the resources directory that will contain the saved data
// selector: xpath to the element
{
  "command_name": "element_screenshot",
  "params": {
    "scale": 2,
    "name": "element.png",
    "selector": "(//img[@class='thumb-image loaded'])[5]",
    "snapshot_name": "s1"
  }
}
```
```json
// Collects metadata on html elements
// snapshot_name: the subfolder name in the resources directory that will contain the saved data
// selector: the element of which to extract nodes from
{
  "command_name": "collect_nodes",
  "params": {
    "selector":"body",
    "wait_ready": false,
    "snapshot_name": "s1",
    "recurse": true,
    "prepopulate": true,
    "get_styles": true
  }
}
```
```json
// Clicks on a element
// query_type: 
  // search: search by xpath
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
// snapshot_name: the subfolder name in the resources directory that will contain the saved data
{
  "command_name": "save_html", 
  "params": {
    "snapshot_name": "s1"
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

```json
// collects snapshots of all versions of the html page over a fixed period of time
// stops when the iteration is complete, or when the a repeat in the html is hit
// iter_limit: the maximum iterations that a page scan will occur
// pause_time: the time to sleep between iterations in milliseconds
// snapshot_name: the name of the snapshot directories that will be generated
// starting_snapshot: the snapshot folder number to start with: <snapshot_name>_<starting_snapshot>
// save_html: whether to save a html page of the current snapshot
// save_node: whether to save a html page of the current node data
// save_full_page_image: whether to save a screenshot of the current html page 
{
  "iter_limit": 10,
  "pause_time": 5000,
  "starting_snapshot": 1,
  "snapshot_name": "snapshot",
  "save_html": true,
  "save_node": true,
  "save_full_page_image": true
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
            "snapshot_name": "s1"
          }
        },
        {
          "command_name": "element_screenshot",
          "params": {
            "scale":2,
            "name": "element.png",
            "selector": "(//img[@class='thumb-image loaded'])[5]",
            "snapshot_name": "s1"
          }
        },
        {
          "command_name": "collect_nodes",
          "params": {
            "selector":"body",
            "wait_ready": false,
            "snapshot_name": "s1"
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
            "snapshot_name": "s1"
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

To execute the agent on a base64 json string run
```shell
agent run -b eyJvcGVyYXRpb...
```


