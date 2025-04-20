import {useState} from 'react';
import './App.css';
import {CloudUpload, GoogleAuth, UserAuthCode} from "../wailsjs/go/main/App";
import {CircleHelp} from "lucide-react"
import {BrowserOpenURL} from "../wailsjs/runtime";

function App() {
    const [filePath, setFilePath] = useState<string>('');
    const [worldName, setWorldName] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);
    const [showCode, setShowCode] = useState<boolean>(false);
    const [userCode, setUserCode] = useState<string>('');
    const updateFiles = () => setLoading(false);

    const Upload = (e: any, worldName: string) => {
      e.preventDefault();
      setLoading(true);
      CloudUpload(worldName).then(updateFiles);
    }

    const handleAuth = () => {
      GoogleAuth().then((url: string) => {
        BrowserOpenURL(url);
        setShowCode(true);
      })
    }

    const handleCode = () => {
      if (userCode !== null && userCode.length > 0) {
        UserAuthCode(userCode)
        setShowCode(false);
      } else {
        console.error("User code is empty");
      }
    }

    const buttonClass = "className= border text-zinc-500 rounded-md px-4 py-2 cursor-pointer transition duration-300 hover:text-zinc-50"

    // UserAuthCode
    return (
        <div id="App">
          <div className="flex flex-col gap-5 justify-center items-center">
            <h1 className="font-bold text-4xl">MINEVCS</h1>
            <form className="flex gap-8 items-center justify-center flex-col" onSubmit={(e) => Upload(e, worldName)}>
            <div className="flex justify-center items-start flex-col gap-2">
                <div className="flex gap-2 items-center justify-center">
                  <label htmlFor="file-path">Minecraft Save Path:</label>
                  <CircleHelp size={15}/>
                </div>
                <input type="text" placeholder="UserHomeDir/Library/Application Support/minecraft/saves/" id="file-path" value={filePath} onChange={(e) => setFilePath(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
              </div>
              <div className="flex justify-center items-start flex-col gap-2">
                <label htmlFor="world-name">Minecraft World Name:</label>
                <input type="text" placeholder="WORLDNAME" id="world-name" autoFocus value={worldName} onChange={(e) => setWorldName(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
              </div>
              <div className="flex justify-center items-start gap-2">
                <button type="button" className={buttonClass} onClick={handleAuth}>Auth</button>
                <button type="submit" className={buttonClass}>Upload</button>
              </div>
            </form>
            {showCode && (
              <div className="flex justify-center items-center gap-2">
                <input type="text" placeholder="4/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" value={userCode} onChange={(e) => setUserCode(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
                <button onClick={handleCode} className={buttonClass}>Login</button>
              </div>
            )}
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
