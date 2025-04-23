import {useState, useEffect} from 'react';
import './App.css';
import {GoogleAuth, UserAuthCode, CheckIfAuthenticated, SaveUserData, GetUserData} from "../wailsjs/go/main/App";
import {CircleHelp, Info} from "lucide-react"
import {BrowserOpenURL, EventsOn} from "../wailsjs/runtime";
import {Link} from "react-router-dom";

function Home() {
    const [minecraftSavePath, setMinecraftSavePath] = useState<string>('');
    const [minecraftLauncherPath, setMinecraftLauncherPath] = useState<string>('');
    const [worldName, setWorldName] = useState<string>('');
    const [showCode, setShowCode] = useState<boolean>(false);
    const [userCode, setUserCode] = useState<string>('');
    const [showTooltip, setShowTooltip] = useState<string | null>(null);
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
    const [logs, setLogs] = useState<string[]>([]);
    const userOS = window.navigator.platform;
    console.log(userOS);
    const isMac = userOS.startsWith("Mac");
    const defaultMinecraftLauncherPath = isMac ? "/Applications/Minecraft.app/Contents/MacOS/launcher" : "C:\\Program Files (x86)\\Minecraft Launcher\\MinecraftLauncher.exe";
    const defaultMinecraftSavePath = isMac ? "/Library/Application Support/minecraft/saves/" : "C:\\Users\\%username%\\AppData\\Roaming\\.minecraft\\saves\\";
 
    useEffect(() => {
      CheckIfAuthenticated().then((isAuth: boolean) => {
        setIsAuthenticated(isAuth);
      });

      GetUserData().then((data) => {
        if (data) {
          setMinecraftLauncherPath(data.minecraftLauncher);
          setMinecraftSavePath(data.minecraftDirectory);
          setWorldName(data.worldName);
        } else {
          console.log("User hasn't set their data yet")
        }
      });     

      const off = EventsOn("log", (msg) => {
        setLogs((prev) => [...prev.slice(-199), msg as string]);
      });
    
      return () => {
        off();
      };
    }, []);

    useEffect(() => {
      const logsElement = document.getElementById('logs');
      if (logsElement) {
        logsElement.scrollTop = logsElement.scrollHeight;
      }
    }, [logs]);

    const handleAuth = () => {
      GoogleAuth().then((url: string) => {
        BrowserOpenURL(url);
        setShowCode(true);
      })
    }

    const verifyCode = () => {
      if (userCode !== null && userCode.length > 0) {
        UserAuthCode(userCode)
        setShowCode(false);
        setIsAuthenticated(true);
      } else {
        console.error("User code is empty");
      }
    }

    const setSavePath = (path: string) => {
      setMinecraftSavePath(path);
    }

    const saveUserSettings = () => {
      if (!minecraftSavePath || !worldName || !minecraftLauncherPath) return;
      SaveUserData(minecraftLauncherPath, minecraftSavePath, worldName).then(() => {
        console.log("User settings saved successfully", minecraftSavePath, worldName);
      });
    }

    const getButtonClass = (disabled: boolean) =>
    `border text-zinc-500 rounded-md px-4 py-2 transition duration-300 ${
      disabled
        ? 'opacity-50 cursor-not-allowed'
        : 'cursor-pointer hover:text-zinc-50'
    }`

    return (
      <div className="flex flex-col gap-5 justify-center items-center relative py-6">
        <Link to="/about" className="absolute top-0 right-0 mx-5 my-6 cursor-pointer opacity-75 hover:opacity-100 transition duration-300">
            <Info size={20}/>
        </Link>
        <h1 className="font-bold text-4xl">MINEVCS</h1>
        {!isAuthenticated ? (
          <>
            <p>Please authorize Google Drive access</p>
            <button type="button" className={getButtonClass(false)} onClick={handleAuth}>Auth</button>
            {showCode && (
              <div className="flex justify-center items-center gap-2">
                <input type="text" placeholder="4/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" value={userCode} onChange={(e) => setUserCode(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
                <button onClick={verifyCode} className={getButtonClass(!userCode)}>Login</button>
              </div>
            )}
          </>
        ) : (
          <>
            <form className="flex gap-8 items-center justify-center flex-col">
            <div className="flex justify-center items-start flex-col gap-2">
              <div className="flex gap-2 items-center justify-center relative">
                  <label htmlFor="file-path">Minecraft Exe Path:</label>
                  <CircleHelp size={15} onMouseEnter={() => setShowTooltip('launch')} onMouseLeave={() => setShowTooltip(null)}/>
                  {showTooltip === 'launch' && (
                    <div className="absolute top-6 left-0 bg-white text-black p-2 rounded-md shadow-md z-10">
                      <p className="text-xs">You can find your Launcher exe path in your finder.</p>
                      <ol className='list-decimal list-inside'>
                        <li className="text-xs">Find your Minecraft App</li>
                        <li className="text-xs">Right click and click "Show Package Contents"</li>
                        <li className="text-xs">Inside "Contents" find your OS folder and finally copy paste the file path to the <pre>launcher.exe</pre></li>
                        <li className="text-xs">For example: <pre>/Application/Minecraft.app/Contents/MacOS/launcher</pre></li>
                      </ol>
                      </div>)
                  }
              </div>
              <div className="flex gap-2 items-center justify-center">
                <input type="text" placeholder="/Applications/Minecraft.app/Contents/MacOS/launcher" id="file-path" value={minecraftLauncherPath} onChange={(e) => setMinecraftLauncherPath(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>              
                <button onClick={() => setMinecraftLauncherPath(defaultMinecraftLauncherPath)} className={getButtonClass(false)} type="button">Default</button>
              </div>
                <div className="flex gap-2 items-center justify-center relative">
                  <label htmlFor="file-path">Minecraft Save Path:</label>
                  <CircleHelp size={15} onMouseEnter={() => setShowTooltip('save')} onMouseLeave={() => setShowTooltip(null)}/>
                  {showTooltip === 'save' && (
                    <div className="absolute top-6 left-0 bg-white text-black p-2 rounded-md shadow-md z-10">
                      <p className="text-xs">You can find your save path in the Minecraft Launcher.</p>
                      <ol className='list-decimal list-inside'>
                        <li className="text-xs">Open Minecraft Launcher</li>
                        <li className="text-xs">Click "Installations"</li>
                        <li className="text-xs">Click the folder icon next to the installation you want to use</li>
                        <li className="text-xs">Go to your folder labeled "saves"</li>
                        <li className="text-xs">Copy the path from the address bar starting from <pre>/Library/</pre></li>
                        <li className="text-xs">For example: <pre>/Library/Application Support/minecraft/saves/</pre></li>
                      </ol>
                      </div>)
                  }
                </div>
                <div className="flex justify-center items-center gap-2">
                  <input type="text" placeholder="/Library/Application Support/minecraft/saves/" id="file-path" value={minecraftSavePath} onChange={(e) => setSavePath(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
                  <button onClick={() => setMinecraftSavePath(defaultMinecraftSavePath)} className={getButtonClass(false)} type="button">Default</button>
                </div>
              </div>
              <div className={`flex justify-center items-start flex-col gap-2 w-full`}>
                <label htmlFor="world-name">Minecraft World To Sync:</label>
                <input type="text" placeholder="World Name" id="world-name" value={worldName} onChange={(e) => setWorldName(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
              </div>
              <div className="flex justify-center items-start gap-2">
                <button type="button" className={getButtonClass(!worldName || !minecraftSavePath || !minecraftLauncherPath || !isAuthenticated)} disabled={!minecraftLauncherPath || !worldName || !minecraftSavePath || !isAuthenticated} onClick={saveUserSettings}>Save Settings</button>
              </div>
            </form>
            <div className="flex justify-center items-center flex-col w-full">
            <div className="flex justify-start items-start outline flex-col w-full h-80 overflow-y-scroll" id='logs'>
              {logs.map((log, index) => (
                <pre key={index} className="text-sm">{log}</pre>
              ))}
            </div>
          </div>
          </>
        )}
        <p className={`text-zinc-100 text-sm mt-1 ${isAuthenticated ? '' : 'fixed bottom-0 mb-6'}`}>© {new Date().getFullYear()} MineVCS. All rights reserved.</p>
      </div>
    )
}

export default Home
