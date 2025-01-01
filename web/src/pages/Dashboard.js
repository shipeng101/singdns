import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  IconButton,
  Button,
  CircularProgress,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { getSystemStatus } from '../services/api';
import PageHeader from '../components/PageHeader';

const Dashboard = ({ mode, setMode, onDashboardClick }) => {
  // ... component code ...
  return (
    <Box>
      <PageHeader title="仪表盘" />
      {/* ... rest of the JSX ... */}
    </Box>
  );
};

export default Dashboard; 