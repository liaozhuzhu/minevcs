const Logs = ({logs} : {logs: string[]}) => {
    return (
        <div className="flex justify-start items-start outline flex-col w-2/3 h-screen overflow-y-scroll gap-2" id='logs'>
            {logs.map((log: string, index: number) => (
            <pre key={index} className="text-xs">{log}</pre>
            ))}
        </div>
    )
}

export default Logs;