# Guide to the NeuroScript `tree` Toolset

The `tree` toolset provides a powerful way to work with structured data, like JSON, in a more flexible and granular manner. Instead of just parsing and printing text, it transforms data into a navigable graph of interconnected nodes. This allows for precise inspection, modification, and querying of complex data structures.

---

## 1. Core Concepts

### The Tree Structure
When you load a JSON string, it's converted into a tree. This isn't just a text blob; it's a collection of individual nodes held in memory.

* **Tree Handle**: When a tree is created (e.g., from `LoadJSON`), you get back a unique **handle** (e.g., `"GenericTree::xyz-abc"`). This handle is your reference to the entire tree structure for all subsequent operations.
* **Nodes**: Every element in the original data—every object, array, string, number, etc.—becomes a **node**. Each node has a unique ID within the tree (e.g., `"node-1"`, `"node-2"`).
* **Relationships**: Nodes are linked. An object node has attributes that point to other nodes. An array node has a list of child node IDs. Every node (except the root) knows its parent's ID.

### Node Properties
Every node in the tree has the following properties, which you can view with `GetNode`:

* `id`: The unique identifier for this node (e.g., `"node-5"`).
* `type`: The data type of the node. This can be `object`, `array`, `string`, `number`, `boolean`, or `null`. If the source JSON object has a `"type"` field (e.g., `{"type": "file", "name": "a.txt"}`), that value is used as the node's type.
* `value`: The actual data for simple nodes (e.g., `"hello world"`, `123`, `true`). For `object` and `array` nodes, this is `nil`.
* `attributes`: A map of key-value pairs.
    * For **object nodes**, the keys are the JSON object keys, and the values are the **node IDs** of the corresponding child nodes.
    * For **all nodes**, this map is also used to store user-defined **metadata** (see `SetNodeMetadata`).
* `children`: A list of child node IDs. This is primarily used by `array` nodes.
* `parent_id`: The ID of the node's parent. This is empty for the root node.

---

## 2. Getting Started: Loading and Inspecting

The most common workflow starts with loading data.

```ns
// 1. Load JSON data into a tree and get its handle.
let json_data = "{\"project\": {\"name\": \"NeuroScript\", \"version\": 0.6}, \"active\": true}"
let tree_handle = tool.Tree.LoadJSON(json_data)

// 2. Get the root node of the tree.
let root_node = tool.Tree.GetRoot(tree_handle)

// 3. Inspect the root node's properties.
// The root_node variable now holds a map, like:
// {
//   "id": "node-1",
//   "type": "object",
//   "value": nil,
//   "attributes": {"project": "node-2", "active": "node-5"},
//   "children": [],
//   "parent_id": ""
// }

// 4. Get a specific child node using its ID from the parent's attributes.
let project_node_id = root_node.attributes.project
let project_node = tool.Tree.GetNode(tree_handle, project_node_id)

// 5. You can also get a node directly by its path from the root.
let version_node = tool.Tree.GetNodeByPath(tree_handle, "project.version")
// version_node.value will be 0.6
```

---

## 3. Tool Reference

### Loading and Saving

* **LoadJSON(json_string)**
    * **Description**: Parses a JSON string into a new tree.
    * **Returns**: A string `tree_handle`.
    * **Example**: `let handle = tool.Tree.LoadJSON("{\"a\": 1}")`

* **ToJSON(tree_handle)**
    * **Description**: Converts an entire tree back into a pretty-printed JSON string.
    * **Returns**: A JSON string.
    * **Example**: `let json_out = tool.Tree.ToJSON(handle)`

* **RenderText(tree_handle)**
    * **Description**: Creates a human-readable, indented text diagram of the tree structure. Useful for debugging.
    * **Returns**: A multi-line string.
    * **Example**: `print(tool.Tree.RenderText(handle))`

### Navigation

* **GetRoot(tree_handle)**
    * **Description**: Retrieves the top-level node of the tree.
    * **Returns**: A map representing the root node.

* **GetNode(tree_handle, node_id)**
    * **Description**: Retrieves any node by its unique ID.
    * **Returns**: A map representing the node.

* **GetNodeByPath(tree_handle, path)**
    * **Description**: Retrieves a node using a dot-separated path from the root (e.g., `key.0.name`).
    * **Returns**: A map representing the found node.

* **GetParent(tree_handle, node_id)**
    * **Description**: Gets the parent of the specified node.
    * **Returns**: A map representing the parent node, or `nil` if the node is the root.

* **GetChildren(tree_handle, node_id)**
    * **Description**: Gets the children of an **array** node. Fails if the node is not of type `array`.
    * **Returns**: A list of child node IDs.

### Modification

* **AddChildNode(handle, parent_id, id_suggestion, type, value, key)**
    * **Description**: Adds a new node to a parent.
    * **Returns**: The ID of the new node.
    * **Example**: `let new_id = tool.Tree.AddChildNode(h, "node-2", "user_profile", "object", nil, "profile")`

* **RemoveNode(tree_handle, node_id)**
    * **Description**: Removes a node and all its descendants from the tree.

* **SetValue(tree_handle, node_id, new_value)**
    * **Description**: Changes the `value` of a simple node (string, number, etc.). Cannot be used on `object` or `array` nodes.

### Attributes & Metadata

* **SetObjectAttribute(handle, object_node_id, key, child_node_id)**
    * **Description**: Links an existing node as a child of an `object` node under a specific key. This is the primary way to build relationships.

* **RemoveObjectAttribute(handle, object_node_id, key)**
    * **Description**: Unlinks a child from an `object` node but does not delete the child itself.

* **SetNodeMetadata(handle, node_id, key, value)**
    * **Description**: Attaches a simple string key-value pair to *any* node's `attributes` map. This is for storing metadata (like status, comments, etc.) that isn't part of the core tree structure.

* **GetNodeMetadata(handle, node_id)**
    * **Description**: Retrieves all attributes of a node, including structural links and metadata.
    * **Returns**: A map of all attributes.

* **RemoveNodeMetadata(handle, node_id, key)**
    * **Description**: Removes a single metadata key-value pair from a node.

### Querying

* **FindNodes(handle, start_node_id, query_map, max_depth, max_results)**
    * **Description**: Finds all nodes descending from `start_node_id` that match the criteria in `query_map`.
    * **Returns**: A list of matching node IDs.
    * **Example**: `let files = tool.Tree.FindNodes(h, root_id, {"type": "file"}, -1, -1)`

---

## 4. Practical Workflow: Building a Tree

You don't have to start with JSON. You can build a tree from scratch.

```ns
// 1. Start with an empty object as the root.
let handle = tool.Tree.LoadJSON("{}")
let root_id = tool.Tree.GetRoot(handle).id

// 2. Add a new node for a list of users.
let users_list_id = tool.Tree.AddChildNode(handle, root_id, "users", "array", nil, "users")

// 3. Add a user object to the users list.
let user1_id = tool.Tree.AddChildNode(handle, users_list_id, "user1", "object", nil, "")

// 4. Add name and email nodes for the new user.
let name_id = tool.Tree.AddChildNode(handle, user1_id, "", "string", "Alice", "name")
let email_id = tool.Tree.AddChildNode(handle, user1_id, "", "string", "alice@example.com", "email")

// 5. Add some metadata to the user node.
tool.Tree.SetNodeMetadata(handle, user1_id, "status", "active")

// 6. View the result.
print(tool.Tree.ToJSON(handle))
// Output will be:
// {
//   "users": [
//     {
//       "email": "alice@example.com",
//       "name": "Alice",
//       "status": "active"
//     }
//   ]
// }
