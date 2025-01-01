import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  IconButton,
  FormControlLabel,
  Switch,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Grid,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { getNodes, updateNode, deleteNode } from '../services/api';
import PageHeader from '../components/PageHeader';

// ... rest of the code ... 