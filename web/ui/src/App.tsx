import React from 'react';
import { Button } from 'reactstrap';
import Navigation from './Navigation';
import {
  BrowserRouter,
  Routes,
  Route,
} from "react-router-dom";
import './App.css';
import Query from "./Query";
import Documentation from "./Documentation";
import API from "./API";

function App() {
  return (
    <BrowserRouter>
    <div>
      <div>
      <Navigation />
      </div>
      <Routes>
        <Route path="/" element={<Button color="danger">Danger!</Button>} />
        <Route path="query" element={<Query />} />
        <Route path="documentation" element={<Documentation />} />
        <Route path="api" element={<API />} />
      </Routes>
    </div>
    </BrowserRouter>
  );
}

export default App;
