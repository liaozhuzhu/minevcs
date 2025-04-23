import { Link } from "react-router-dom";

const About = () => {
    return (
        <div className="flex flex-col gap-6 justify-center items-center text-center p-6 max-w-4xl mx-auto">
            <h1 className="font-bold text-4xl text-zinc-100">ABOUT MINEVCS</h1>
            <div className="flex flex-col gap-2">
                <h2 className="text-2xl font-semibold text-zinc-100">Incentive</h2>
                <p className="text-zinc-100 text-lg text-left">
                    MineVCS was born out of personal need — I had two laptops: one for gaming at home, and another I used while out and about. Minecraft Java Edition, unlike Bedrock, doesn’t offer native cloud syncing. I wanted a seamless way to continue my worlds across devices.
                </p>
            </div>

            <div className="flex flex-col gap-2">
                <h2 className="text-2xl font-semibold text-zinc-100">How It Works</h2>
                <p className="text-zinc-100 text-lg text-left">
                    MineVCS syncs your Minecraft Java world folder to your personal <strong className="text-green-400">Google Drive</strong>, mimicking cloud saving. When Minecraft launches, MineVCS pulls your latest save. When you exit the game (via <span className="text-gray-500">"Quit Game"</span>), it uploads to your Drive — keeping everything in sync across devices.
                </p>
                <br/>
                <p className="text-zinc-100 text-lg text-left">
                    For this system to function correctly, it requires:
                    <ul className="list-disc list-inside text-left mt-2">
                        <li>Accurate path to your Minecraft launcher</li>
                        <li>The full save directory path (ending in <code className="text-blue-300">/saves/</code>)</li>
                        <li>A successful token authorization via <span className="text-green-400">Google</span></li>
                    </ul>
                    These details are essential — so be sure to enter them carefully. Incorrect or missing info will prevent syncing from working as expected and may lead to data loss or corruption.
                </p>
                <br/>
                <p className="text-zinc-100 text-sm mt-1">
                    Configuration and tokens are stored locally in a hidden folder per device. Please do not modify these files manually to avoid sync issues or broken authentication.
                </p>
            </div>

            <div className="flex flex-col gap-2">
                <h2 className="text-2xl font-semibold text-zinc-100">Code</h2>
                <p className="text-zinc-100 text-lg text-left">
                    The code for minevcs is publicly available and can be found at <a href="https://github.com/liaozhuzhu/minevcs" target="_blank" className="text-blue-400 underline">github.com/liaozhuzhu/minevcs</a>
                </p>
            </div>
            <Link to="/" className="text-blue-500 underline transition duration-300 hover:text-blue-300"> Back to Home</Link>
            <p className="text-zinc-100 text-sm mt-1">© {new Date().getFullYear()} MineVCS. All rights reserved.</p>
        </div>
    );
};

export default About;
