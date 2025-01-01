import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Grid,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { getRules, updateRule, deleteRule } from '../services/api';
import PageHeader from '../components/PageHeader';

const Rules = ({ mode, setMode, onDashboardClick }) => {
  // ... component code ...
  return (
    <Box>
      <PageHeader title="规则管理" />
      {/* ... rest of the JSX ... */}
    </Box>
  );
};

export default Rules; 