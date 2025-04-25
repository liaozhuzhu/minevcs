const LaunchTooltip = () => {
    return (
        <div className="absolute top-6 left-0 bg-white text-black p-2 rounded-md shadow-md z-10">
                      <p className="text-xs">You can find your Launcher exe path in your finder.</p>
                      <ol className='list-decimal list-inside'>
                        <li className="text-xs">Find your Minecraft App</li>
                        <li className="text-xs">Right click and click "Show Package Contents"</li>
                        <li className="text-xs">Inside "Contents" find your OS folder and finally copy paste the file path to the <pre>launcher.exe</pre></li>
                        <li className="text-xs">For example: <pre>/Application/Minecraft.app/Contents/MacOS/launcher</pre></li>
                      </ol>
                      </div>
    )
}
export default LaunchTooltip;