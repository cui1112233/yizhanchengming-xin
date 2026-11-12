import React from 'react';
import { createRoot } from 'react-dom/client';
import 'antd/dist/reset.css';
import '../shared/styles/global.css';
import '../shared/styles/theme-schemes.css';
import '../shared/styles/cinematic-glass.css';
import { installClientErrorReporting } from '../shared/error-reporting';
import { AdminApp } from './App';

installClientErrorReporting();

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <AdminApp />
  </React.StrictMode>
);
