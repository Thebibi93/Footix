import { useState, useEffect } from 'react'
import './App.css'

function App() {

  const [message, setMessage] = useState("chargement...")

  useEffect(() => {
    fetch("http://localhost:8080/api/hello")
      .then(res => res.text())
      .then(data => setMessage(data))
  }, [])

  return (
    <>
      <h1>Footix</h1>
      <p>Message du serveur :</p>
      <h2>{message}</h2>
    </>
  )
}

export default App