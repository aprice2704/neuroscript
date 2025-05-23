// src/extension.ts
import * as path from 'path';
import { workspace, ExtensionContext, window, Disposable } from 'vscode'; // Added Disposable
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
    console.log('Activating NeuroScript Language Client...');

    // Path to the server executable (nslsp_executable)
    const serverModule = context.asAbsolutePath(
        path.join('server', 'nslsp_executable') // Ensure 'nslsp_executable' is in 'vscode-neuroscript/server/'
                                                // Adjust if your executable is elsewhere relative to the extension.
    );

    console.log(`LSP Server Executable Path: ${serverModule}`);

    // Server options: tells the client how to run the server
    const serverOptions: ServerOptions = {
        command: serverModule,
        transport: TransportKind.stdio // Use stdio for communication
    };

    // Client options: defines how the client behaves
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'neuroscript' }],
        synchronize: {
            // fileEvents: workspace.createFileSystemWatcher('**/.clientrc') // Optional: if server needs to watch other files
        },
        // Capture stderr from the server for logging/debugging in an output channel
        outputChannel: window.createOutputChannel('NeuroScript LSP')
    };

    // Create the language client
    client = new LanguageClient(
        'neuroscriptLSP',              // ID of the language client (must be unique)
        'NeuroScript Language Server', // Name of the language client (shows in Output channel dropdown)
        serverOptions,
        clientOptions
    );

    // Start the client.
    // The start() method returns a Disposable that should be pushed to context.subscriptions.
    // This ensures that the client is properly shutdown when the extension is deactivated.
    console.log('Starting NeuroScript Language Client...');
    context.subscriptions.push(client.start());

    // Optional: If you need to perform actions specifically after the client is ready
    // (i.e., successfully initialized with the server).
    client.onReady().then(() => {
        console.log('NeuroScript Language Client is ready and connected to the server!');
        // You could, for example, send a custom notification to the server here if needed.
    }).catch((e: any) => { // Explicitly type 'e' (error) as 'any' or 'Error'
        console.error('NeuroScript Language Client failed to become ready:', e);
        window.showErrorMessage('NeuroScript Server failed to initialize. Check the "NeuroScript LSP" Output channel.');
    });
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    console.log('Deactivating NeuroScript Language Client...');
    return client.stop();
}