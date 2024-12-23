import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Typography,
  IconButton,
  Stack,
  CircularProgress,
  LinearProgress,
  useTheme,
  Button,
  Chip,
} from '@mui/material';
import {
  OpenInNew as OpenInNewIcon,
  LightMode as LightModeIcon,
  DarkMode as DarkModeIcon,
  Sync as SyncIcon,
} from '@mui/icons-material';
import {
  getSystemInfo,
  getSystemStatus,
  getTrafficStats,
  getRealtimeTraffic,
  startService,
  stopService,
  getSubscriptions,
  refreshSubscription,
} from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Dashboard = ({ mode, setMode, onDashboardClick = () => {} }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [manualLoading, setManualLoading] = useState(false);
  const [services, setServices] = useState([
    {
      id: 'singbox',
      name: 'Singbox',
      description: '代理服务',
      active: false,
    },
    {
      id: 'mosdns',
      name: 'Mosdns',
      description: 'DNS服务',
      active: false,
    },
  ]);
  const [systemInfo, setSystemInfo] = useState({
    hostname: '',
    platform: '',
    arch: '',
    uptime: 0,
    cpu_usage: 0,
    memory_total: 0,
    memory_used: 0,
    networkUpload: 0,
    networkDownload: 0,
    connections: 0,
    singbox_version: '',
    singbox_uptime: 0,
    mosdns_version: '',
    mosdns_uptime: 0,
  });

  // 定义订阅列表
  const [subscriptions, setSubscriptions] = useState([]);

  // Define fetchData with useCallback
  const fetchData = useCallback(async () => {
    try {
      const [sysInfo, sysStatus] = await Promise.all([
        getSystemInfo(),
        getSystemStatus(),
      ]);

      // Update system info
      setSystemInfo({
        hostname: sysInfo.hostname || '',
        platform: sysInfo.system?.platform || '',
        arch: sysInfo.system?.arch || '',
        uptime: sysInfo.uptime || 0,
        cpu_usage: sysInfo.system?.cpu?.usage || 0,
        memory_total: sysInfo.system?.memory?.total || 0,
        memory_used: sysInfo.system?.memory?.used || 0,
        networkUpload: sysInfo.system?.network?.tx_rate || 0,
        networkDownload: sysInfo.system?.network?.rx_rate || 0,
        connections: sysInfo.system?.network?.connections || 0,
        singbox_version: sysStatus.services?.singbox_version || '-',
        singbox_uptime: sysStatus.services?.singbox_uptime || 0,
        mosdns_version: sysStatus.services?.mosdns_version || '-',
        mosdns_uptime: sysStatus.services?.mosdns_uptime || 0,
      });

      // Update services status
      setServices(prevServices => prevServices.map(service => ({
        ...service,
        active: sysStatus.services?.find(s => s.name === service.id)?.is_running || false,
      })));

      // Update subscriptions
      setSubscriptions(sysStatus.subscriptions?.map(sub => ({
        id: sub.id,
        name: sub.name || '未命名订阅',
        nodeCount: sub.node_count || 0,
        lastUpdate: formatDate(sub.last_update),
        expireTime: formatDate(sub.expire_time),
      })) || []);

    } catch (error) {
      console.error('Failed to fetch data:', error);
    }
  }, []);

  // 添加自动刷新的 useEffect
  useEffect(() => {
    // 首次加载数据
    fetchData();

    // 设置定时器，每5秒刷新一次数据
    const intervalId = setInterval(fetchData, 5000);

    // 清理函数
    return () => {
      clearInterval(intervalId);
    };
  }, [fetchData]);

  const handleServiceToggle = async (serviceId, action = 'toggle') => {
    try {
      setManualLoading(true);
      const service = services.find(s => s.id === serviceId);
      if (service) {
        if (action === 'restart') {
          await stopService(serviceId);
          await new Promise(resolve => setTimeout(resolve, 1000)); // 等待1秒
          await startService(serviceId);
        } else if (action === 'stop' || service.active) {
          await stopService(serviceId);
        } else {
          await startService(serviceId);
        }
      }
      await fetchData();
    } catch (error) {
      console.error(`Failed to ${action} ${serviceId}:`, error);
    } finally {
      setManualLoading(false);
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  const formatUptime = (seconds) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}天${hours}小时${minutes}分钟`;
  };

  const formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // 手动刷新函数
  const handleManualRefresh = async () => {
    setManualLoading(true);
    await fetchData();
    setManualLoading(false);
  };

  // 修改订阅更新函数
  const handleUpdateAllSubscriptions = async () => {
    try {
      setManualLoading(true);
      const subs = await getSubscriptions();
      await Promise.all(subs.map(sub => refreshSubscription(sub.id)));
      await fetchData();
    } catch (error) {
      console.error('Failed to update all subscriptions:', error);
    } finally {
      setManualLoading(false);
    }
  };

  const handleUpdateSubscription = async (id) => {
    try {
      setManualLoading(true);
      await refreshSubscription(id);
      await fetchData();
    } catch (error) {
      console.error(`Failed to update subscription ${id}:`, error);
    } finally {
      setManualLoading(false);
    }
  };

  return (
    <Box>
      <Card sx={styles.headerCard}>
        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" alignItems="center" spacing={2}>
              <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
                仪表盘
              </Typography>
              {manualLoading && (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
              )}
            </Stack>
            <Stack direction="row" spacing={1}>
              <IconButton
                size="small"
                onClick={handleManualRefresh}
                disabled={manualLoading}
                sx={styles.iconButton}
              >
                <SyncIcon fontSize="small" />
              </IconButton>
              {onDashboardClick && (
                <IconButton
                  size="small"
                  onClick={onDashboardClick}
                  sx={styles.iconButton}
                >
                  <OpenInNewIcon fontSize="small" />
                </IconButton>
              )}
              <IconButton
                size="small"
                onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
                sx={styles.iconButton}
              >
                {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
              </IconButton>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
      {manualLoading && <LinearProgress sx={{ mb: 1 }} />}

      <Grid container spacing={2} sx={{ mt: 0.5 }}>
        {/* 第一行：系信息、CPU使用率、内存使用率、网络状态 */}
        <Grid item xs={12} sm={6} md={3} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={1}>
                <Typography variant="h6" sx={{ fontWeight: 500 }}>系统信息</Typography>
                <Stack spacing={0.5}>
                  <Typography variant="body2" color="text.secondary">
                    操作系统：{systemInfo.platform}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    架构：{systemInfo.arch}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    运行时间：{formatUptime(systemInfo.uptime)}
                  </Typography>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={2}>
                <Typography variant="h6" sx={{ fontWeight: 500 }}>CPU 使用率</Typography>
                <Box sx={{ position: 'relative', display: 'inline-flex', justifyContent: 'center', width: '100%' }}>
                  <CircularProgress
                    variant="determinate"
                    value={systemInfo.cpu_usage}
                    size={100}
                    thickness={4}
                    sx={{
                      color: (theme) =>
                        systemInfo.cpu_usage > 80 ? theme.palette.error.main :
                        systemInfo.cpu_usage > 60 ? theme.palette.warning.main :
                        theme.palette.success.main
                    }}
                  />
                  <Box sx={{
                    position: 'absolute',
                    top: '50%',
                    left: '50%',
                    transform: 'translate(-50%, -50%)',
                    textAlign: 'center'
                  }}>
                    <Typography variant="h5" sx={{ fontWeight: 500 }}>
                      {Math.round(systemInfo.cpu_usage)}%
                    </Typography>
                  </Box>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={2}>
                <Typography variant="h6" sx={{ fontWeight: 500 }}>内存使用率</Typography>
                <Box sx={{ position: 'relative', display: 'inline-flex', justifyContent: 'center', width: '100%' }}>
                  <CircularProgress
                    variant="determinate"
                    value={(systemInfo.memory_used / systemInfo.memory_total) * 100}
                    size={100}
                    thickness={4}
                    sx={{
                      color: (theme) =>
                        (systemInfo.memory_used / systemInfo.memory_total) * 100 > 80 ? theme.palette.error.main :
                        (systemInfo.memory_used / systemInfo.memory_total) * 100 > 60 ? theme.palette.warning.main :
                        theme.palette.success.main
                    }}
                  />
                  <Box sx={{
                    position: 'absolute',
                    top: '50%',
                    left: '50%',
                    transform: 'translate(-50%, -50%)',
                    textAlign: 'center'
                  }}>
                    <Typography variant="h5" sx={{ fontWeight: 500 }}>
                      {Math.round((systemInfo.memory_used / systemInfo.memory_total) * 100)}%
                    </Typography>
                  </Box>
                </Box>
                <Typography variant="body2" color="text.secondary" align="center">
                  {formatBytes(systemInfo.memory_used)} / {formatBytes(systemInfo.memory_total)}
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={1}>
                <Typography variant="h6" sx={{ fontWeight: 500 }}>网络状态</Typography>
                <Stack spacing={0.5}>
                  <Typography variant="body2" color="text.secondary">
                    上传：{formatBytes(systemInfo.networkUpload)}/s
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    下载：{formatBytes(systemInfo.networkDownload)}/s
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    连接数：{systemInfo.connections}
                  </Typography>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 第二行：服务状态卡片 */}
        <Grid item xs={12} sm={6} md={4} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={2}>
                <Stack direction="row" alignItems="center" spacing={1}>
                  <Typography variant="h6" sx={{ fontWeight: 500 }}>Singbox 状态</Typography>
                  <Chip
                    label={services[0].active ? "运行中" : "已停止"}
                    size="small"
                    sx={{
                      ...services[0].active ? styles.chip.success : styles.chip.error,
                      background: services[0].active 
                        ? 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)'
                        : 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    }}
                  />
                </Stack>
                <Stack spacing={1}>
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Typography variant="body2" color="text.secondary">版本</Typography>
                    <Typography variant="body2">{systemInfo.singbox_version || '-'}</Typography>
                  </Stack>
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Typography variant="body2" color="text.secondary">运行时间</Typography>
                    <Typography variant="body2">{formatUptime(systemInfo.singbox_uptime)}</Typography>
                  </Stack>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={4} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={2}>
                <Stack direction="row" alignItems="center" spacing={1}>
                  <Typography variant="h6" sx={{ fontWeight: 500 }}>Mosdns 状态</Typography>
                  <Chip
                    label={services[1].active ? "运行中" : "已停止"}
                    size="small"
                    sx={{
                      ...services[1].active ? styles.chip.success : styles.chip.error,
                      background: services[1].active 
                        ? 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)'
                        : 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    }}
                  />
                </Stack>
                <Stack spacing={1}>
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Typography variant="body2" color="text.secondary">版本</Typography>
                    <Typography variant="body2">{systemInfo.mosdns_version || '-'}</Typography>
                  </Stack>
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Typography variant="body2" color="text.secondary">运行时间</Typography>
                    <Typography variant="body2">{formatUptime(systemInfo.mosdns_uptime)}</Typography>
                  </Stack>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={12} md={4} sx={{ display: 'flex' }}>
          <Card sx={{ ...styles.card, width: '100%', height: '100%' }}>
            <CardContent>
              <Stack spacing={2}>
                <Typography variant="h6" sx={{ fontWeight: 500 }}>服务控制</Typography>
                <Stack spacing={2}>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      Singbox 服务
                    </Typography>
                    <Stack direction="row" spacing={1}>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={manualLoading || services[0].active}
                        onClick={() => handleServiceToggle('singbox', 'start')}
                        sx={{
                          flex: 1,
                          bgcolor: 'success.main',
                          '&:hover': {
                            bgcolor: 'success.dark',
                          },
                          '&:disabled': {
                            bgcolor: 'success.light',
                            opacity: 0.5,
                          },
                        }}
                      >
                        启动
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={manualLoading || !services[0].active}
                        onClick={() => handleServiceToggle('singbox', 'restart')}
                        sx={{
                          flex: 1,
                          bgcolor: 'warning.main',
                          '&:hover': {
                            bgcolor: 'warning.dark',
                          },
                          '&:disabled': {
                            bgcolor: 'warning.light',
                            opacity: 0.5,
                          },
                        }}
                      >
                        重启
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={manualLoading || !services[0].active}
                        onClick={() => handleServiceToggle('singbox', 'stop')}
                        sx={{
                          flex: 1,
                          bgcolor: 'error.main',
                          '&:hover': {
                            bgcolor: 'error.dark',
                          },
                          '&:disabled': {
                            bgcolor: 'error.light',
                            opacity: 0.5,
                          },
                        }}
                      >
                        停止
                      </Button>
                    </Stack>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      Mosdns 服务
                    </Typography>
                    <Stack direction="row" spacing={1}>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={manualLoading || services[1].active}
                        onClick={() => handleServiceToggle('mosdns', 'start')}
                        sx={{
                          flex: 1,
                          bgcolor: 'success.main',
                          '&:hover': {
                            bgcolor: 'success.dark',
                          },
                          '&:disabled': {
                            bgcolor: 'success.light',
                            opacity: 0.5,
                          },
                        }}
                      >
                        启动
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={manualLoading || !services[1].active}
                        onClick={() => handleServiceToggle('mosdns', 'restart')}
                        sx={{
                          flex: 1,
                          bgcolor: 'warning.main',
                          '&:hover': {
                            bgcolor: 'warning.dark',
                          },
                          '&:disabled': {
                            bgcolor: 'warning.light',
                            opacity: 0.5,
                          },
                        }}
                      >
                        重启
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={manualLoading || !services[1].active}
                        onClick={() => handleServiceToggle('mosdns', 'stop')}
                        sx={{
                          flex: 1,
                          bgcolor: 'error.main',
                          '&:hover': {
                            bgcolor: 'error.dark',
                          },
                          '&:disabled': {
                            bgcolor: 'error.light',
                            opacity: 0.5,
                          },
                        }}
                      >
                        停止
                      </Button>
                    </Stack>
                  </Box>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 第三行：订阅信息 */}
        <Grid item xs={12}>
          <Card sx={{ ...styles.card, width: '100%' }}>
            <CardContent>
              <Stack spacing={2}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Typography variant="h6" sx={{ fontWeight: 500 }}>订阅信息</Typography>
                  <Button
                    variant="contained"
                    size="small"
                    startIcon={<SyncIcon />}
                    onClick={handleUpdateAllSubscriptions}
                    disabled={manualLoading}
                    sx={styles.actionButton}
                  >
                    更新全部
                  </Button>
                </Box>
                <Grid container spacing={2}>
                  {subscriptions.map((subscription) => (
                    <Grid item xs={12} sm={6} md={4} key={subscription.id}>
                      <Card sx={styles.subscriptionCard}>
                        <CardContent>
                          <Stack spacing={1}>
                            <Stack direction="row" justifyContent="space-between" alignItems="center">
                              <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
                                {subscription.name}
                              </Typography>
                              <Button
                                variant="outlined"
                                size="small"
                                startIcon={<SyncIcon />}
                                onClick={() => handleUpdateSubscription(subscription.id)}
                                disabled={manualLoading}
                                sx={styles.outlinedButton}
                              >
                                更新
                              </Button>
                            </Stack>
                            <Stack spacing={0.5}>
                              <Typography variant="body2" color="text.secondary">
                                节点数量：{subscription.nodeCount}
                              </Typography>
                              <Typography variant="body2" color="text.secondary">
                                上次更新：{subscription.lastUpdate}
                              </Typography>
                              <Typography variant="body2" color="text.secondary">
                                过期时间：{subscription.expireTime}
                              </Typography>
                            </Stack>
                          </Stack>
                        </CardContent>
                      </Card>
                    </Grid>
                  ))}
                </Grid>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard; 