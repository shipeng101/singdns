import React, { useState, useEffect } from 'react';
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
  Switch,
  Chip,
} from '@mui/material';
import {
  OpenInNew as OpenInNewIcon,
  LightMode as LightModeIcon,
  DarkMode as DarkModeIcon,
  Sync as SyncIcon,
} from '@mui/icons-material';
import { getSystemInfo, getSystemStatus, startService, stopService } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Dashboard = ({ mode, setMode, onDashboardClick = () => {} }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [loading, setLoading] = useState(false);
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
  });

  // 定义服务列表
  const services = [
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
  ];

  // 定义订阅列表
  const [subscriptions, setSubscriptions] = useState([]);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000); // 每5秒更新一次
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [sysInfo, sysStatus] = await Promise.all([
        getSystemInfo(),
        getSystemStatus(),
      ]);

      // 更新系统信息
      setSystemInfo({
        hostname: sysInfo.hostname || '',
        platform: sysInfo.platform || '',
        arch: sysInfo.arch || '',
        uptime: sysInfo.uptime || 0,
        cpu_usage: sysInfo.cpu_usage || 0,
        memory_total: sysInfo.memory_total || 0,
        memory_used: sysInfo.memory_used || 0,
        networkUpload: sysInfo.network_upload || 0,
        networkDownload: sysInfo.network_download || 0,
        connections: sysInfo.connections || 0,
      });

      // 更新服务状态
      services[0].active = sysStatus.services?.singbox || false;
      services[1].active = sysStatus.services?.mosdns || false;

      // 更新订阅信息
      setSubscriptions(sysStatus.subscriptions?.map(sub => ({
        id: sub.id,
        name: sub.name || '未命名订阅',
        nodeCount: sub.node_count || 0,
        lastUpdate: formatDate(sub.last_update),
        expireTime: formatDate(sub.expire_time),
      })) || []);

    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleServiceToggle = async (serviceId) => {
    try {
      setLoading(true);
      const service = services.find(s => s.id === serviceId);
      if (service) {
        if (service.active) {
          await stopService(serviceId);
        } else {
          await startService(serviceId);
        }
      }
      await fetchData();
    } catch (error) {
      console.error(`Failed to toggle ${serviceId}:`, error);
    } finally {
      setLoading(false);
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
    const days = Math.floor(seconds / (24 * 60 * 60));
    const hours = Math.floor((seconds % (24 * 60 * 60)) / (60 * 60));
    const minutes = Math.floor((seconds % (60 * 60)) / 60);
    return `${days}天 ${hours}小时 ${minutes}分钟`;
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

  const handleUpdateAllSubscriptions = async () => {
    try {
      setLoading(true);
      // TODO: 实现更新所有订阅的 API 调用
      await fetchData();
    } catch (error) {
      console.error('Failed to update all subscriptions:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateSubscription = async (id) => {
    try {
      setLoading(true);
      // TODO: 实现更新单个订阅的 API 调用
      await fetchData();
    } catch (error) {
      console.error(`Failed to update subscription ${id}:`, error);
    } finally {
      setLoading(false);
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
              {loading && (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
              )}
            </Stack>
            <Stack direction="row" spacing={1}>
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

      {loading && <LinearProgress sx={{ mb: 1 }} />}

      <Grid container spacing={1.5}>
        {/* 第一行：系统信息、CPU使用率、内存使用率、网络状态 */}
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={styles.card}>
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

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={styles.card}>
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

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={styles.card}>
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

        <Grid item xs={12} sm={6} md={3}>
          <Card sx={styles.card}>
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
        <Grid item xs={12} sm={6} md={4}>
          <Card sx={styles.card}>
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
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={4}>
          <Card sx={styles.card}>
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
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={12} md={4}>
          <Card sx={styles.card}>
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
                        disabled={loading || services[0].active}
                        onClick={() => handleServiceToggle('singbox')}
                        sx={{
                          flex: 1,
                          background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                          color: '#fff',
                          '&:disabled': {
                            background: 'linear-gradient(45deg, rgba(76, 175, 80, 0.5) 30%, rgba(129, 199, 132, 0.5) 90%)',
                            color: 'rgba(255, 255, 255, 0.5)',
                          },
                        }}
                      >
                        启动
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={loading || !services[0].active}
                        onClick={() => handleServiceToggle('singbox')}
                        sx={{
                          flex: 1,
                          background: 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                          color: '#fff',
                          '&:disabled': {
                            background: 'linear-gradient(45deg, rgba(244, 67, 54, 0.5) 30%, rgba(229, 115, 115, 0.5) 90%)',
                            color: 'rgba(255, 255, 255, 0.5)',
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
                        disabled={loading || services[1].active}
                        onClick={() => handleServiceToggle('mosdns')}
                        sx={{
                          flex: 1,
                          background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                          color: '#fff',
                          '&:disabled': {
                            background: 'linear-gradient(45deg, rgba(76, 175, 80, 0.5) 30%, rgba(129, 199, 132, 0.5) 90%)',
                            color: 'rgba(255, 255, 255, 0.5)',
                          },
                        }}
                      >
                        启动
                      </Button>
                      <Button
                        variant="contained"
                        size="small"
                        disabled={loading || !services[1].active}
                        onClick={() => handleServiceToggle('mosdns')}
                        sx={{
                          flex: 1,
                          background: 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                          color: '#fff',
                          '&:disabled': {
                            background: 'linear-gradient(45deg, rgba(244, 67, 54, 0.5) 30%, rgba(229, 115, 115, 0.5) 90%)',
                            color: 'rgba(255, 255, 255, 0.5)',
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
          <Card sx={styles.card}>
            <CardContent>
              <Stack spacing={2}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Typography variant="h6" sx={{ fontWeight: 500 }}>订阅信息</Typography>
                  <Button
                    variant="contained"
                    size="small"
                    startIcon={<SyncIcon />}
                    onClick={handleUpdateAllSubscriptions}
                    disabled={loading}
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
                                disabled={loading}
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