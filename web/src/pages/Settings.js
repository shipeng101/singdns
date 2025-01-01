import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  FormControl,
  FormControlLabel,
  RadioGroup,
  Radio,
} from '@mui/material';
import { getSettings, updateSettings } from '../services/api';
import PageHeader from '../components/PageHeader';

// ... rest of the code ... 