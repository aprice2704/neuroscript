# NeuroScript Setup for GoLand

This guide provides the steps to configure GoLand for NeuroScript development, including syntax highlighting and integration with the NeuroScript Language Server (nslsp).

Because GoLand's settings are not designed to be modified by external scripts, this is a one-time manual setup process.

### Step 1: Install the Language Server

First, ensure the nslsp executable is built and available in your system's PATH. You can do this by running the following command from the root of the NeuroScript repository:

sh make install 

### Step 2: Define the NeuroScript File Type

Next, teach GoLand about .ns files.

1. Go to Settings/Preferences > Editor > File Types.

2. Under Recognized File Types, click the + button to add a new type.

3. In the Add File Type dialog, enter the following:

   * Name: NeuroScript

   * Description: NeuroScript Language File

4. In the Keywords tab for your new NeuroScript file type, click the + icon and add the following keywords one by one:

   * Line Comment: -- (and add # and // as well)

   * Keywords 1: func means returns needs optional on error do endon event

   * Keywords 2: emit must mustbe fail clear_error clear_event ask into

   * Keywords 3: if else endif while endwhile for each in endfor

   * Keywords 4: break continue call return set and or not no some

5. With the NeuroScript type still selected, go to the File name patterns section and click +. Add *.ns and *.ns.txt to associate these patterns with the file type.

6. Click Apply.

### Step 3: Configure the LSP Server

Finally, tell GoLand how to communicate with the language server.

1. Go to Settings/Preferences > Languages & Frameworks > Language Server Protocol.

2. Click the + button to add a new server configuration.

3. In the LSP Server settings, configure the following:

   * Remote Server / Local Executable: Choose Local Executable.

   * Executable: Enter the full path to your nslsp executable. You can find this by running which nslsp in your terminal.

   * Arguments: Leave this blank.

   * Working Directory: Leave this blank.

   * File types: Select the NeuroScript file type you created in the previous step.

4. Click Apply and OK.

After completing these steps, restart GoLand. When you open a .ns or .ns.txt file, you should now have syntax highlighting and receive real-time diagnostics from nslsp.