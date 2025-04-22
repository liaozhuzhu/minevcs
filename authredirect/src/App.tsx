import { useEffect, useState } from 'react'
import './App.css'
import { useSearchParams } from 'react-router-dom'

function App() {
  const [code, setCode] = useState<string>('')
  const [searchParams, setSearchParams] = useSearchParams()
  const [copied, setCopied] = useState<boolean>(false)

  useEffect(() => {
    const code = searchParams.get('code')
    if (code) {
      setCode(code)
    } else {
      console.error('No code found in URL')
    }
  }
  , [searchParams])

  if (!code || code.length === 0) {
    return <h1>MINEVCS REDIRECT</h1>
  }

  return (
    <div style={{height: '100vh', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'start', overscrollBehavior: 'none', gap: '10px'}}>
      {copied && (
        <div style={{position: 'absolute', top: '10px', right: '10px', backgroundColor: 'green', color: 'white', padding: '10px', borderRadius: '5px'}}>
          Copied to clipboard!
        </div>
      )}
      <img src="/appicon.png" style={{width: '150px'}}/>
      <p style={{textAlign: 'center'}}>Copy the following code and paste it into the respective box in minevcs application</p>
      <div style={{display: 'flex', flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: '5px'}}>
        <pre style={{fontSize: '1rem', outline: '1px solid black', padding: '8px', borderRadius: '10px'}} id="codeText">{code}</pre>
        <button onClick={() => {
          const codeText = document.getElementById('codeText')
          if (codeText) {
            navigator.clipboard.writeText(codeText.innerText)
          }
          setCopied(true)
          setTimeout(() => {
            setCopied(false)
          }
          , 2000)
        }} className="btn">Copy</button>
      </div>
    </div>
  )
}

export default App
