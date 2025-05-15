const LaunchTooltip = () => {
    return (
        <div className="absolute top-6 left-0 bg-white text-black p-2 rounded-md shadow-md z-10">
                      <p className="text-xs">You can find your Launcher exe path in your find / file explorer</p>
                      <ol className='list-decimal list-inside'>
                        <li className="text-xs">Find your Minecraft Launcher App (most likely in Applications)</li>
                        <li className="text-xs">For MacOS, right click and click "Show Package Contents"</li>
                        <li className="text-xs">For Windows, your exe might be contained wihtin "Minecraft Launcher"</li>
                        <li className="text-xs">Go into "Contents" and if on MacOS go into "MacOS"</li>
                        <li className="text-xs">Finally copy paste the file path to the <pre>launcher(.exe)</pre></li>
                        <li className="text-xs">For example: <pre>/Application/Minecraft.app/Contents/MacOS/launcher</pre></li>
                      </ol>
                      </div>
    )
}
export default LaunchTooltip;