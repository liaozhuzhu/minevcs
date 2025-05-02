import {useState, useEffect} from 'react';
import './App.css';
import {GoogleAuth, UserAuthCode, CheckIfAuthenticated, SaveUserData, GetUserData, PushIfAhead} from "../wailsjs/go/main/App";
import { CircleHelp, Settings, Info } from 'lucide-react';
import {BrowserOpenURL, EventsOn} from "../wailsjs/runtime";
import {Link} from "react-router-dom";
import SaveTooltip from './components/SaveTooltip';
import LaunchTooltip from './components/LaunchTooltip';
import Logs from './components/Logs';

function Home() {
    const [minecraftSavePath, setMinecraftSavePath] = useState<string>('');
    const [minecraftLauncherPath, setMinecraftLauncherPath] = useState<string>('');
    const [authError, setAuthError] = useState<string | null>(null);
    const [worldName, setWorldName] = useState<string>('');
    const [showCode, setShowCode] = useState<boolean>(false);
    const [userCode, setUserCode] = useState<string>('');
    const [showTooltip, setShowTooltip] = useState<string | null>(null);
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
    const [logs, setLogs] = useState<string[]>([]);
    const userOS = window.navigator.platform;
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

    useEffect(() => {
      if (!minecraftSavePath || !worldName) return;
      PushIfAhead().then(() => {
        console.log("Checking if local is ahead of remote");
      }).catch((error) => {
        console.error("Error pushing if ahead", error);
      });
    }, [minecraftSavePath, worldName]);

    const handleAuth = () => {
      GoogleAuth().then((url: string) => {
        BrowserOpenURL(url);
        setShowCode(true);
      })
    }

    const verifyCode = () => {
      if (userCode !== null && userCode.length > 0) {
        UserAuthCode(userCode)
          .then(() => {
            setShowCode(false);
            setIsAuthenticated(true);
          })
          .catch((error) => {
            setAuthError(error);
          });
      } else {
        console.error("User code is empty");
      }
    }

    const setSavePath = (path: string) => {
      setMinecraftSavePath(path);
    }

    const saveUserSettings = (e: any) => {
        e.preventDefault();
        if (!minecraftSavePath || !worldName || !minecraftLauncherPath) return;
        SaveUserData(minecraftLauncherPath, minecraftSavePath, worldName).then(() => {
            console.log("User settings saved successfully", minecraftSavePath, worldName);
        });
    }

    const getButtonClass = (disabled: boolean) =>
    `border text-zinc-500 rounded-md px-4 py-2 transition duration-300 ${
      disabled
        ? 'opacity-80 cursor-not-allowed'
        : 'cursor-pointer hover:text-zinc-50'
    }`

    return (
      <div className="flex flex-col justify-center items-center h-screen">
        <Link to="/about" className="absolute bottom-0 left-0 m-5 cursor-pointer opacity-75 hover:opacity-100 transition duration-300 group">
            <Info size={25} className="transition duration-300 group-hover:rotate-360"/>
        </Link>
        {!isAuthenticated ? (
            <div className="flex justify-start items-center flex-col gap-2 shadow-xl w-[500px] p-4 rounded-3xl bg-zinc-800">
                <img src="/logo.png" alt="MineVCS Logo" className="w-20 h-20"/>
                <h1 className="text-2xl font-bold text-zinc-50">MineVCS</h1>
                <p className="text-lg">{showCode ? 'Please Enter Your Code' : 'Please Authorize Google Drive access'}</p>
                {showCode && (
                <div className="flex justify-center items-center gap-2 mt-10">
                    <input 
                    type="text" 
                    placeholder="4/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" 
                    value={userCode} onChange={(e) => setUserCode(e.target.value)} 
                    autoFocus
                    className="shadow-xl bg-zinc-50 focus:ring-0 focus:outline-none rounded-md text-xs placeholder:opacity-50 px-2 py-3 w-80 text-zinc-900"
                    />
                    <button onClick={verifyCode} className={getButtonClass(!userCode)}>Verify</button>
                </div>
                )}
                <div className="my-10 cursor-pointer text-zinc-900 opacity-90 hover:opacity-100 bg-zinc-50 rounded-xl px-4 py-2 transition duration-300 flex justify-between items-center gap-3" onClick={handleAuth}>
                    <img src="/drive.png" alt="Google Logo" className="w-8 h-8"/>
                    <p className="text-sm">Authorize Redirect</p>
                </div>
                {authError && (
                    <p className="text-red-500 text-xs">{authError}</p>
                )}
            </div>
        )
         : (
          <div className="flex justify-between items-start w-full h-screen">
            <form className="flex gap-8 items-center justify-center flex-col w-1/2 h-screen" onSubmit={(e) => saveUserSettings(e)}>
                <div className="flex justify-center items-start flex-col gap-8">
                    <div className="flex flex-col gap-2 items-start justify-center">
                        <div className="flex gap-2 items-center justify-center relative">
                            <label htmlFor="file-path">Minecraft Exe Path:</label>
                            <CircleHelp size={15} onMouseEnter={() => setShowTooltip('launch')} onMouseLeave={() => setShowTooltip(null)}/>
                            {showTooltip === 'launch' && (<LaunchTooltip/>)}
                        </div>
                        <div className="flex gap-2 items-start justify-center flex-col ">
                            <input type="text" 
                                placeholder="/Applications/Minecraft.app/Contents/MacOS/launcher" 
                                id="file-path" 
                                value={minecraftLauncherPath} 
                                onChange={(e) => setMinecraftLauncherPath(e.target.value)} 
                                className="border border-zinc-50 focus:ring-0 focus:outline-none rounded-md text-xs placeholder:opacity-50 px-2 py-3 w-80 bg-zinc-900 text-zinc-100"/>              
                            <p onClick={() => setMinecraftLauncherPath(defaultMinecraftLauncherPath)} className="cursor-pointer underline text-blue-400 hover:text-blue-500 transition duration-300 text-xs">Default</p>
                        </div>
                    </div>
                    <div className="flex flex-col gap-2 items-start justify-center">
                        <div className="flex gap-2 items-center justify-center relative">
                            <label htmlFor="file-path">Minecraft Save Path:</label>
                            <CircleHelp size={15} onMouseEnter={() => setShowTooltip('save')} onMouseLeave={() => setShowTooltip(null)}/>
                            {showTooltip === 'save' && (<SaveTooltip/>)}
                        </div>
                        <div className="flex justify-center items-start gap-2 flex-col">
                            <input type="text" 
                                placeholder="/Library/Application Support/minecraft/saves/" 
                                id="file-path" 
                                value={minecraftSavePath} 
                                onChange={(e) => setSavePath(e.target.value)} 
                                className="border border-zinc-50 focus:ring-0 focus:outline-none rounded-md text-xs placeholder:opacity-50 px-2 py-3 w-80 bg-zinc-900 text-zinc-100"
                                />
                            <p onClick={() => setMinecraftSavePath(defaultMinecraftSavePath)} className="underline text-blue-400 hover:text-blue-500 transition duration-300 cursor-pointer text-xs">Default</p>
                        </div>
                    </div>
                    <div className="flex flex-col gap-2 items-start justify-center">
                        <div className="flex gap-2 items-center justify-center relative">
                            <label htmlFor="world-name">Minecraft World To Sync:</label>
                        </div>
                        <div className="flex justify-center items-center gap-2">
                            <input type="text" placeholder="World Name" id="world-name" value={worldName} onChange={(e) => setWorldName(e.target.value)} className="w-80 border border-zinc-50 focus:ring-0 focus:outline-none rounded-md text-xs placeholder:opacity-50 px-2 py-3 bg-zinc-900 text-zinc-100"/>
                        </div>
                    </div>
                </div>
                <button type="submit" 
                    className={getButtonClass(!worldName || !minecraftSavePath || !minecraftLauncherPath || !isAuthenticated) + ' group flex justify-center items-center gap-2 text-xs'} 
                    disabled={!minecraftLauncherPath || !worldName || !minecraftSavePath || !isAuthenticated} 
                >
                    <span className="transition-transform duration-300 group-hover:rotate-45"><Settings/></span>
                    Save Settings
                </button>
            </form>
            <Logs logs={logs}/>
          </div>
        )}
      </div>
    )
}

export default Home
