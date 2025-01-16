import React from 'react';
import { Routes, Route } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Nodes from './pages/Nodes';
import Rules from './pages/Rules';
import Settings from './pages/Settings';
import Subscriptions from './pages/Subscriptions';

const AppRoutes = () => {
  return (
    <Routes>
      <Route path="/" element={<Dashboard />} />
      <Route path="/nodes" element={<Nodes />} />
      <Route path="/rules" element={<Rules />} />
      <Route path="/settings" element={<Settings />} />
      <Route path="/subscriptions" element={<Subscriptions />} />
    </Routes>
  );
};

export default AppRoutes; 