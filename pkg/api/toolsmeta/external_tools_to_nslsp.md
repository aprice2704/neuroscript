That's a good point. A command-line argument makes the exporter program much more flexible and is standard practice for a tool like this.

I've updated the instruction file with a new version of the exporter program that uses a command-line flag to specify the output file.

Here is the updated content for `/home/aprice/dev/neuroscript/pkg/api/toolsmeta/external_tools_to_nslsp.md`.

-----

### Integrating Custom Tools with the NeuroScript LSP

To get IDE features like autocompletion and error-checking for your project's custom NeuroScript tools, you need to provide metadata about them to the NeuroScript Language Server (`nslsp`). This is a straightforward, three-step process using a dedicated Go program within your project.

-----

### 1\. Create a Tool Exporter Program

In your project (e.g., FDM), create a new Go program responsible for generating the tool metadata. A good location for this is `tools/meta_exporter/main.go`.

This program will:

1.  **Blank-import** all of your project's custom tool packages. This is the most critical step, as it triggers their `init()` functions and registers the tools, making them discoverable.
2.  Import and call the `toolsmeta.ExportTools` function from the NeuroScript API.
3.  Use the standard `flag` package to accept a command-line argument for the output file path.

Here is a complete template for your `tools/meta_exporter/main.go`:

```go
// filename: tools/meta_exporter/main.go
package main

import (
    "flag"
    "fmt"
    "log"

    // 1. Import your project's custom tool packages using a blank identifier.
    // Replace these with the actual import paths for your FDM tools.
    _ "github.com/your-org/fdm/pkg/neuroscript/tools/core"
    _ "github.com/your-org/fdm/pkg/neuroscript/tools/graph"

    // 2. Import the NeuroScript tools metadata API.
    "github.com/aprice2704/neuroscript/pkg/api/toolsmeta"
)

func main() {
    // 3. Define a command-line flag for the output file path.
    outputFile := flag.String("o", "./tools/fdm-tools.json", "Output file for the JSON metadata")
    flag.Parse()

    // 4. Call the exporter function.
    // It discovers all registered tools (including yours) and
    // writes their definitions to the specified output file.
    if err := toolsmeta.ExportTools(*outputFile); err != nil {
        log.Fatalf("Error exporting custom tool metadata: %v", err)
    }

    fmt.Printf("Successfully exported tool metadata to %s\n", *outputFile)
}
```

-----

### 2\. Generate the Metadata File

Run your new exporter program from your project's root directory, providing the output path via the `-o` flag.

```bash
go run ./tools/meta_exporter/main.go -o ./tools/fdm-tools.json
```

You should integrate this command into your project's `Makefile` to ensure the metadata is kept up-to-date whenever your tools change.

-----

### 3\. Configure Your Workspace

Finally, tell the `nslsp` where to find your new metadata file. Create or edit the `.vscode/settings.json` file in your project's root and add the `nslsp.externalToolMetadata` setting.

```json
// filename: .vscode/settings.json
{
  "nslsp.externalToolMetadata": [
    "./tools/fdm-tools.json"
  ]
}
```

You can add multiple paths to this array if your project uses several different toolsets.
