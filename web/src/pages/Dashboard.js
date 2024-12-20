import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  IconButton,
  Button,
  CircularProgress,
  useTheme,
  Snackbar,
  Alert,
  Stack
} from '@mui/material';
import {
  Memory as CPUIcon,
  Storage as MemoryIcon,
  Speed as SpeedIcon,
  Refresh as RefreshIcon,
  OpenInNew as OpenInNewIcon
} from '@mui/icons-material';
import * as api from '../services/api';

const ServiceCard = ({ title, version, uptime, icon: Icon }) => {
  const theme = useTheme();
  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box display="flex" alignItems="center" mb={2}>
          <Icon sx={{ mr: 1, color: theme.palette.primary.main }} />
          <Typography variant="h6">{title}</Typography>
        </Box>
        <Stack spacing={1}>
          <Typography variant="body2" color="text.secondary">
            版本: {version || '获取中...'}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            运行时间: {uptime || '获取中...'}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
};

const StatusCard = ({ title, value, total, icon: Icon, unit = '%' }) => {
  const theme = useTheme();

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box display="flex" alignItems="center" mb={2}>
          <Icon sx={{ mr: 1, color: theme.palette.primary.main }} />
          <Typography variant="h6">{title}</Typography>
        </Box>
        <Box display="flex" alignItems="baseline" mb={1}>
          <Typography variant="h4" component="span">
            {typeof value === 'number' ? value.toFixed(1) : '0.0'}
          </Typography>
          <Typography variant="body1" color="text.secondary" ml={1}>
            {unit}
          </Typography>
        </Box>
        {total && (
          <Typography variant="body2" color="text.secondary">
            {value.toFixed(1)} / {total.toFixed(1)} {unit}
          </Typography>
        )}
      </CardContent>
    </Card>
  );
};

const Dashboard = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState(null);
  const [snackbar, setSnackbar] = useState({
    open: false,
    message: '',
    severity: 'error'
  });
  const [retryCount, setRetryCount] = useState(0);

  const fetchStatus = useCallback(async () => {
    try {
      setLoading(true);
      const response = await api.getSystemStatus();
      if (response.data) {
        setStatus(response.data);
        setRetryCount(0);
      }
    } catch (error) {
      console.error('获取系统状态失败:', error);
      setSnackbar({
        open: true,
        message: error.response?.data?.message || '获取系统状态失败',
        severity: 'error'
      });
      // 如果是认证错误，不再重试
      if (error.response?.status !== 401 && retryCount < 3) {
        setRetryCount(prev => prev + 1);
        setTimeout(fetchStatus, 5000);
      }
    } finally {
      setLoading(false);
    }
  }, [retryCount]);

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 10000);
    return () => clearInterval(interval);
  }, [fetchStatus]);

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5">仪表盘</Typography>
        <Box>
          <Button
            variant="outlined"
            startIcon={<OpenInNewIcon />}
            href="/metacubexd/"
            target="_blank"
            sx={{ mr: 1 }}
          >
            MetaCubeXD
          </Button>
          <Button
            variant="outlined"
            startIcon={<OpenInNewIcon />}
            href="/yacd/"
            target="_blank"
            sx={{ mr: 1 }}
          >
            YACD
          </Button>
          <IconButton onClick={fetchStatus} disabled={loading}>
            {loading ? <CircularProgress size={24} /> : <RefreshIcon />}
          </IconButton>
        </Box>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <ServiceCard
            title="Sing-Box"
            version={status?.singbox?.version}
            uptime={status?.singbox?.uptime}
            icon={SpeedIcon}
          />
        </Grid>
        <Grid item xs={12} md={6}>
          <ServiceCard
            title="MosDNS"
            version={status?.mosdns?.version}
            uptime={status?.mosdns?.uptime}
            icon={SpeedIcon}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <StatusCard
            title="CPU 使用率"
            value={status?.cpu?.usage || 0}
            icon={CPUIcon}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <StatusCard
            title="内存使用率"
            value={status?.memory?.used || 0}
            total={status?.memory?.total || 0}
            icon={MemoryIcon}
            unit="GB"
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <StatusCard
            title="网络速度"
            value={status?.network?.speed?.up || 0}
            icon={SpeedIcon}
            unit="MB/s"
          />
        </Grid>
      </Grid>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
      >
        <Alert 
          severity={snackbar.severity}
          onClose={() => setSnackbar({ ...snackbar, open: false })}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Dashboard; 