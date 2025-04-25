const SaveTooltip = () => {
    return (
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
                      </div>
    )
}

export default SaveTooltip;