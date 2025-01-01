import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  IconButton,
  Stack,
  Box,
  CircularProgress,
} from '@mui/material';
import {
  LightMode as LightModeIcon,
  DarkMode as DarkModeIcon,
  Sync as SyncIcon,
  Dashboard as DashboardIcon,
} from '@mui/icons-material';

const PageHeader = ({
  title,
  mode,
  setMode,
  loading,
  onRefresh,
  onDashboardClick,
  styles,
}) => {
  return (
    <Card sx={styles.headerCard}>
      <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Stack direction="row" alignItems="center" spacing={2}>
            <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
              {title}
            </Typography>
            {loading && (
              <Box sx={{ display: 'flex', alignItems: 'center' }}>
                <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
              </Box>
            )}
          </Stack>
          <Stack direction="row" spacing={1}>
            {onRefresh && (
              <IconButton
                size="small"
                onClick={onRefresh}
                disabled={loading}
                sx={styles.iconButton}
                title="刷新"
              >
                <SyncIcon fontSize="small" />
              </IconButton>
            )}
            {onDashboardClick && (
              <IconButton
                size="small"
                onClick={onDashboardClick}
                sx={styles.iconButton}
                title="打开仪表盘"
              >
                <DashboardIcon fontSize="small" />
              </IconButton>
            )}
            <IconButton
              size="small"
              onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
              sx={styles.iconButton}
              title={mode === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}
            >
              {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
            </IconButton>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
};

export default PageHeader; 