import {useState} from 'react';
import './App.css';
import {ListFiles} from "../wailsjs/go/main/App";
import {CircleHelp} from "lucide-react"

function App() {
    const [filePath, setFilePath] = useState<string>('');
    const [worldName, setWorldName] = useState<string>('');
    const [files, setFiles] = useState<string[]>([]);
    const updateFiles = (files: string[]) => setFiles(files);

    const listFiles = (e: any, worldName: string) => {
      e.preventDefault();
      ListFiles(worldName).then(updateFiles);
    }

    return (
        <div id="App">
          <div className="flex flex-col gap-5 justify-center items-center">
            <h1 className="font-bold text-4xl">MINEVCS</h1>
            <form className="flex gap-8 items-center justify-center flex-col" onSubmit={(e) => listFiles(e, worldName)}>
            <div className="flex justify-center items-start flex-col gap-2">
                <div className="flex gap-2 items-center justify-center">
                  <label htmlFor="file-path">Minecraft Save Path:</label>
                  <CircleHelp/>
                </div>
                <input type="text" placeholder="UserHomeDir/Library/Application Support/minecraft/saves/" id="file-path" value={filePath} onChange={(e) => setFilePath(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
              </div>
              <div className="flex justify-center items-start flex-col gap-2">
                <label htmlFor="world-name">Minecraft World Name:</label>
                <input type="text" placeholder="WORLDNAME" id="world-name" autoFocus value={worldName} onChange={(e) => setWorldName(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
              </div>
              <button type="submit" className="bg-blue-500 text-white rounded-md px-4 py-2 hover:cursor-pointer">List Files</button>
            </form>
            {files.length > 0 && (
              <div className="flex flex-col gap-2">
                <h2 className="font-bold text-2xl">Files:</h2>
                <ul className="list-disc list-inside">
                  {files.map((file, index) => (
                    <li key={index} className="text-sm">{file}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </div>
    )
}

export default App
