import { Link } from "react-router-dom";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";

const About = () => {
    return (
        <div className="flex flex-col gap-3 justify-center items-center text-center p-6 max-w-4xl mx-auto">
            <h1 className="font-bold text-4xl text-zinc-100">ABOUT MINEVCS</h1>
            <div className="flex flex-col justify-center items-start gap-10 mt-10">
                <div className=" w-full text-start flex flex-col gap-5">
                    <h2 className="text-2xl font-semibold text-zinc-100">Incentive</h2>
                    <p className="text-zinc-100 text-sm">
                        MineVCS was born out of personal need — I had two laptops: one for home, and another while out and about. Minecraft Java Edition, unlike Bedrock, doesn’t offer native cloud syncing. MineVCS was created to bridge this gap, allowing me to seamlessly play on either device without losing progress. It’s a simple solution that I hope will help others facing the same issue.
                    </p>
                </div>

                <div className=" w-full text-start flex flex-col gap-5">
                    <h2 className="text-2xl font-semibold text-zinc-100">How It Works (High Level)</h2>
                    <p className="text-zinc-100 text-sm">
                        MineVCS syncs your Minecraft Java world folder to your personal Google Drive, mimicking cloud saving. When Minecraft launches, MineVCS pulls your latest save. When you exit the game MineVCS uploads to your Drive — keeping everything in sync across devices.
                    </p>
                    <img src="/design.png" alt="design" className="w-full mx-auto" />
                </div>

                <div className=" w-full text-start flex flex-col gap-5">
                    <h2 className="text-2xl font-semibold text-zinc-100">How It Works (Detailed)</h2>
                    <p className="text-zinc-100 text-sm">
                    A first time user will be forced to connect to a chosen Google Drive account, requiring them to go through a custom redirect site (minevcs-redirect.vercel.app) to streamline the OAuth process. Once the user is authenticated, they are able to start configuring their application by selecting the path to their Minecraft launcher and the world directory they wish to sync. 
                    Once these settings are saved, a <code>config</code> file is created in a hidden directory in the user's home directory, allowing the user to easily access and modify their settings in the future as well as allowing the application to run without requiring the user to reconfigure their settings every time they launch the application.
                    </p>
                    <p className="text-zinc-100 text-sm">
                    Upon detecting the Minecraft launcher starting, MineVCS pulls the latest version of the specified world from Google Drive, ensuring that the local version is up to date.
                    </p>
                    <p className="text-zinc-100 text-sm">
                    When the user exits Minecraft, MineVCS first uploads a temporary <code>lockfile</code> so any subsequent reads done on a user's separate machine know an upload is in progress and don't pull. Following that, MineVCS then zips and uploads the updated world folder to Google Drive.
                    </p>
                    <img src="/detail_design.png" alt="detailed design" className="w-full mx-auto" />
                </div>

                <div className=" w-full text-start flex flex-col gap-5">
                    <h2 className="text-2xl font-semibold text-zinc-100">Assumptions / Limitations</h2>
                    <p className="text-zinc-100 text-sm">
                        MineVCS is designed to work with the Minecraft Java Edition and requires a Google Drive account for cloud storage. It assumes that the user has enough storage space on their Google Drive to accommodate their requested world folder.
                    </p>
                    <p className="text-zinc-100 text-sm">
                        MineVCS currently cannot distinguish between two <code>.zip</code> files with the same name in Google Drive. This means that if a user has two worlds with the same name or any files with the same name as a Minecraft world in their Google Drive, MineVCS will not be able to differentiate between them.
                        This could lead to potential data loss if the user is not careful when selecting their world folder. I plan to add hashing to the world folder to prevent this from happening in the near future.
                    </p>
                    <p className="text-zinc-100 text-sm">
                        MineVCS is currently only available for MacOS as of 04/26/2025, but will soon support Windows as well (and since syncing is done through Google Drive, there won't be any slowdowns between MacOS and Window devices 😁)
                    </p>
                    <p className="text-zinc-100 text-sm">
                        MineVCS creates a <code>.minevcs</code> hidden directory in the user's home directory to store the <code>config</code> file as well as other necessary helper files. This is where the user's settings are stored, and it is recommended that users do not modify or delete this directory unless they know what they are doing.
                    </p>
                </div>

                <div className=" w-full text-start flex flex-col gap-5">
                    <h2 className="text-2xl font-semibold text-zinc-100">Privacy</h2>
                    <p className="text-zinc-100 text-sm">
                        MineVCS does not collect any personal data from users. The application only accesses the user's Google Drive account to sync the specified world folder but does not interact with any files outside of Minecraft related files selected by the user within the Google Drive. MineVCS does not store any data on its own servers.
                    </p>
                </div>

                <div className=" w-full text-start flex flex-col gap-5">
                    <h2 className="text-2xl font-semibold text-zinc-100">Code</h2>
                    <p className="text-zinc-100 text-sm"> The code for minevcs is public as I want to be as transparent about your data as possible and can be found at <span onClick={() => BrowserOpenURL("https://github.com/liaozhuzhu/minevcs")} className="text-blue-400 underline cursor-pointer transition duration-300 hover:text-blue-500">github.com/liaozhuzhu/minevcs</span></p>
                </div>
            </div>
            <Link to="/" className="text-blue-500 underline transition duration-300 hover:text-blue-300 mt-10"> Back to Home</Link>
            <p className="text-zinc-100 text-[10px]">© {new Date().getFullYear()} MineVCS. All rights reserved.</p>
        </div>
    );
};

export default About;
