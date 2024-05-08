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
The Browser commands give your agent access to a Browser. You can do things such as scrape the html and 
node data, click on elements, and take screenshots. You can then chain commands together forming an
operation.

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
// Collects snapshots of all versions of the html page over a fixed period of time
// stops when the iteration is complete, or when the a repeat in the html is hit
// iter_limit: the maximum iterations that a page scan will occur
// pause_time: the time to sleep between iterations in milliseconds
// snapshot_name: the name of the snapshot directories that will be generated
// starting_snapshot: the snapshot folder number to start with: <snapshot_name>_<starting_snapshot>
// save_html: whether to save a html page of the current snapshot
// save_node: whether to save a html page of the current node data
// save_full_page_image: whether to save a screenshot of the current html page 
{
  "command_name": "iterate_html",
  "params": {
    "iter_limit": 10,
    "pause_time": 5000,
    "starting_snapshot": 1,
    "snapshot_name": "snapshot",
    "save_html": true,
    "save_node": true,
    "save_full_page_image": true
  }
}
```

### LLM 

LLM commands allow us to make commands to various LLMs. We handle rate limiting and switch too
backup LLMs if the main LLM fails. 

#### Settings

```json
// try_limit: how many times to re-request the LLM after a failure
// timeout: the max amount of time (in milliseconds) the request can run for before cancelling
// max_tokens: the token limit for a response
// llm settings: the configuration of the llms you wish to use, the first 
// item in the list is the first llm that gets tried
// workflow: what workflow configuration to run the llm command in

// calculating max run time in ms: try_limit * len(llm_settings) * timeout 
{
  "settings": {
    "try_limit": 3,
    "timeout": 15000,
    "max_tokens": 300,
    "llm_settings": [...],
    "workflow": {...}
  }
}
```
##### LLM Settings
###### OpenAI
```json
// name: OpenAI
// api_key: your OpenAI apikey
// model: the name of your open ai model. Accepted options below
// temperature: the temperature used when generating responses between -2 and 2
{
  "name": "OpenAI",
  "api_key": "sk-...",
  "model": "gpt-3.5-turbo", 
  "temperature": 1.0
}
```

| model              | gpt-3.5-turbo | gpt-3.5-turbo-0125 | gpt-3.5-turbo-1106 | gpt-4-turbo-2024-04-09 | gpt-4-0125-preview | gpt-4-1106-preview |
|:-------------------|---------------|--------------------|--------------------|------------------------|--------------------|-------------------:|
| Multimodal         | no            | no                 | yes                | yes                    | no                 |                yes |
| Json Mode          | no            | no                 | no                 | yes                    | no                 |                 no |
| Function Calling   | no            | no                 | yes                | yes                    | no                 |                yes |


##### Workflow Settings

```json
// workflow_type only currently accepted option is chat_completion
{
  "workflow": {
    "workflow_type": "chat_completion"
  }
}
```

#### Messages
These are the type of messages you can send to the llm

```json
// A standard chat request
// message_type: standard
// role: can either be system or user
// content: the statement being asked
{
  "message_type": "standard",
  "message": {
    "role": "user",
    "content": "what is bench ai?"
  }
}
```

```json
// A multimodal chat request
// message_type: multimodal
// role: can either be system or user
// content: the statement being asked
// name: a name you can associate to the message
{
  "message_type": "multimodal",
  "message": {
    "role": "user",
    "content": [...],
    "name": "test"
  }
}

// A text content
// type: text
// text: the statement being told to the agent
{
  "type": "text",
  "text": "what is bench ai?"
}

// A text content
// type: image_url
// image_url: 
    // url: a url or base64 encoded bytes of the image
    // detail: a little description of the image
{
  "type": "image_url",
  "image_url": {
    "url": "https://..."
    "detail": "a image of a park bench",
  }
}
```

```json
// An assistant message this should be auto generated by the llm, do not construct your own
// message_type: assistant
// content: the response of the llm if there are no tool calls
// name: a optional name associated with the message
// tool_calls: the tools being called by the assistant
{
  "message_type": "assistant",
  "message":{
    "content": "bench ai is a company that offers tools for digital accessibility compliance",
    "role": "assistant",
    "name": "...",
    "tool_calls": [...]
  }
}

// The tool call invoked by the assistant
// id: the id of the tool call
// type: the type of the tool call, only function is currently supported
// function:
  // name: the name of the function
  // arguments: the arguements being used in the func
{
  "id": "0000-...",
  "type": "function",
  "function": {
    "name": "my_func",
    "arguments": "...",
  }
}
```

```json
// message_type: tool
// role: tool
// content: the response of the assistant
// tool_call_id: the id of the tool call request
{
  "message_type": "tool",
  "message": {
    "role": "assistant",
    "content": "....",
    "tool_call_id": "0000-..."
  }
}
```

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
    },
    {
      "type": "llm",
      "settings": {
        "try_limit": 3,
        "timeout": 15000,
        "max_tokens": 300,
        "llm_settings": [
          {
            "name": "OpenAI",
            "api_key": "...",
            "model": "gpt-3.5-turbo",
            "temperature": 1.0
          }
        ],
        "workflow": {
          "workflow_type": "chat_completion"
        }
      },
      "command_list": [
        {
          "message_type": "standard",
          "message": {
            "role": "user",
            "content": "You are an expert on digital accessibility, and work as an accessibility Auditor. You are currently performing an audit on a pdf."
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




