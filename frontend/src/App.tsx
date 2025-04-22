import {useState, useEffect} from 'react';
import './App.css';
import {GoogleAuth, UserAuthCode, CheckIfAuthenticated, GetWorlds, SaveUserData, GetUserData} from "../wailsjs/go/main/App";
import {CircleHelp} from "lucide-react"
import {BrowserOpenURL} from "../wailsjs/runtime";

function App() {
    const [minecraftSavePath, setMinecraftSavePath] = useState<string>('');
    const [minecraftLauncherPath, setMinecraftLauncherPath] = useState<string>('');
    const [worlds, setWorlds] = useState<string[]>([]);
    const [worldName, setWorldName] = useState<string>('');
    const [showCode, setShowCode] = useState<boolean>(false);
    const [userCode, setUserCode] = useState<string>('');
    const [showTooltip, setShowTooltip] = useState<string | null>(null);
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);

    useEffect(() => {
      CheckIfAuthenticated().then((isAuth: boolean) => {
        setIsAuthenticated(isAuth);
      });

      GetUserData().then((data) => {
        if (data) {
          setMinecraftLauncherPath(data.minecraftLauncher);
          setMinecraftSavePath(data.minecraftDirectory);
          setWorldName(data.worldName);
          GetWorlds(data.minecraftDirectory).then((worlds: string[]) => {
            setWorlds(worlds);
          }
          );
        } else {
          console.log("User hasn't set their data yet")
        }
      });     
    }, []);

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
      if (path.split("/")[path.split("/").length-1] === "saves") {
        GetWorlds(path).then((worlds: string[]) => setWorlds(worlds))
      } else {
        setWorlds([]);
      }
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
        <div id="App">
          <div className="flex flex-col gap-5 justify-center items-center">
            <h1 className="font-bold text-4xl">MINEVCS</h1>
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
                <input type="text" placeholder="/Applications/Minecraft.app/Contents/MacOS/launcher" id="file-path" value={minecraftLauncherPath} onChange={(e) => setMinecraftLauncherPath(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>              
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
                </div>
              </div>
              <div className={`flex justify-center items-start flex-col gap-2 w-full ${!worlds.length ? 'opacity-50 cursor-not-allowed' : ''}`}>
                <label htmlFor="world-name">Minecraft World</label>
                <select id="world-name" value={worldName} onChange={(e) => {setWorldName(e.target.value)}}   className="h-12 appearance-none w-full border border-zinc-600 rounded-md text-white placeholder-white/50 px-3 py-2 text-sm leading-tight focus:outline-none"                >
                  <option value="" disabled>Select a world</option>
                  {worlds.map((world: string, index: number) => (
                    <option key={index} value={world}>{world}</option>
                  ))}
                </select>
              </div>
              <div className="flex justify-center items-start gap-2">
                {!isAuthenticated && <button type="button" className={getButtonClass(false)} onClick={handleAuth}>Auth</button>}
                <button type="button" className={getButtonClass(!worldName || !minecraftSavePath || !minecraftLauncherPath)} disabled={!worldName || !minecraftSavePath} onClick={saveUserSettings}>Save Settings</button>
              </div>
            </form>
            {showCode && (
              <div className="flex justify-center items-center gap-2">
                <input type="text" placeholder="4/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" value={userCode} onChange={(e) => setUserCode(e.target.value)} className="border border-zinc-300 rounded-md text-xs border-transparent focus:border-transparent focus:ring-0 placeholder:opacity-50 px-2 py-3 w-80"/>
                <button onClick={verifyCode} className={getButtonClass(!userCode)}>Login</button>
              </div>
            )}
          </div>
        </div>
    )
}

export default App
