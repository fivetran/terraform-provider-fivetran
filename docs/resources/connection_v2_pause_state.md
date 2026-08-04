---
page_title: "Resource: fivetran_connection_v2_pause_state"
---

# Resource: fivetran_connection_v2_pause_state

~> **Alpha:** This resource is in alpha and its schema and behavior may change without notice. Not recommended for production use.

Manages the paused state of a `fivetran_connection_v2` connection, separately from the connection resource itself. This lets you pause and resume a connection without triggering changes to (or replacement of) the connection's own configuration.

**`fivetran_connection_v2` always creates connections paused.** Add this resource with `paused = false` to start syncing.

If you are importing both resources, import `fivetran_connection_v2` first, then `fivetran_connection_v2_pause_state` — importing the pause state before the connection exists in state forces a replacement plan.

## Example Usage

```hcl
resource "fivetran_connection_v2_pause_state" "pause_state" {
    connection_id = fivetran_connection_v2.connection.id
    paused        = false
}
```

## Schema

### Required

- `connection_id` (String) The unique identifier for the connection whose pause state is managed. Changing this forces the resource to be replaced.
- `paused` (Boolean) Whether the connection should be paused.

### Read-Only

- `id` (String) The unique identifier for this pause-state resource. This is the connection ID.
