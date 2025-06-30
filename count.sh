#!/bin/bash

# This script runs cloc to count lines of code, then calculates a
# project completion estimate based on the ratio of Go lines to Markdown lines.

# Run cloc and pipe its output to be processed.
# The `tee /dev/tty` command prints the cloc output directly to the terminal
# so you can see it, while also passing the output down the pipe for processing.
cloc \
    --exclude-ext=html,js,class,java \
    --exclude-lang=JSON \
    --exclude-dir=site,tmp,vscode-neuroscript,vim-neuroscript \
    * | tee /dev/tty | {
        # Use awk to process the entire cloc output in one go.
        # This is more efficient than running multiple grep and awk commands.
        awk '
            # For the line starting with "Go ", store the 5th field (code lines)
            /^Go / { lines_go = $5 }

            # For the line starting with "Markdown ", store the 5th field
            /^Markdown / { lines_markdown = $5 }

            # The END block is executed after all lines of input have been processed.
            END {
                # Check if we found values for both Go and Markdown, and Markdown is not zero.
                if (lines_go > 0 && lines_markdown > 0) {
                    # Calculate percentage: (go_lines / (4 * markdown_lines)) * 100
                    percentage = (lines_go * 100) / (4 * lines_markdown)

                    # Print a separator for clarity
                    print "-------------------------------------------------------------------------------"
                    # Print the final formatted result
                    printf "Estimated %% done (Go / (4 * Markdown)): %.2f%%\n", percentage
                    print "-------------------------------------------------------------------------------"

                } else {
                    # If we couldn"t find the lines or Markdown lines are 0, print a message.
                    print "-------------------------------------------------------------------------------"
                    print "Could not calculate percentage. Ensure Go & Markdown files exist."
                    print "-------------------------------------------------------------------------------"
                }
            }
        '
}
