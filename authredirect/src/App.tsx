import { useEffect, useState } from 'react'
import './App.css'
import { useSearchParams } from 'react-router-dom'

function App() {
  const [code, setCode] = useState<string>('')
  const [searchParams, setSearchParams] = useSearchParams()

  useEffect(() => {
    const code = searchParams.get('code')
    if (code) {
      setCode(code)
      setSearchParams({})
    } else {
      console.error('No code found in URL')
    }
  }
  , [searchParams])

  return (
    <div>
      <p>Copy the following code and paste it into the respective box in minevcs application</p>
      <pre>{code}</pre>
    </div>
  )
}

export default App
