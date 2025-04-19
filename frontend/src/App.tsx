import {useState} from 'react';
import './App.css';
import {CloudUpload} from "../wailsjs/go/main/App";
import {CircleHelp} from "lucide-react"

function App() {
    const [filePath, setFilePath] = useState<string>('');
    const [worldName, setWorldName] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);
    const updateFiles = () => setLoading(false);

    const Upload = (e: any, worldName: string) => {
      e.preventDefault();
      setLoading(true);
      CloudUpload(worldName).then(updateFiles);
    }

    return (
        <div id="App">
          <div className="flex flex-col gap-5 justify-center items-center">
            <h1 className="font-bold text-4xl">MINEVCS</h1>
            <form className="flex gap-8 items-center justify-center flex-col" onSubmit={(e) => Upload(e, worldName)}>
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
              <button type="submit" className="bg-blue-500 text-white rounded-md px-4 py-2 hover:cursor-pointer">Upload</button>
            </form>
            {loading && 
              <div className="flex justify-center items-center flex-col gap-2">
                <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-500"></div>
                <p className="text-blue-500">Uploading...</p>
              </div>
            }
          </div>
        </div>
    )
}

export default App
