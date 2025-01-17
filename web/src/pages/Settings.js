import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  IconButton,
  Stack,
  Alert,
  Snackbar,
  CircularProgress,
  LinearProgress,
  useTheme,
  FormControl,
  FormControlLabel,
  RadioGroup,
  Radio,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Switch,
  Select,
  MenuItem,
  Grid,
} from '@mui/material';
import {
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Key as KeyIcon,
  Sync as SyncIcon,
  Dashboard as DashboardIcon,
} from '@mui/icons-material';
import { getCommonStyles } from '../styles/commonStyles';
import { getSettings, updateSettings, updatePassword, updateDNSSettings, generateConfig } from '../services/api';
import axios from 'axios';
import config from '../config';

// 配置 axios 默认值
axios.defaults.baseURL = config.apiBaseUrl;

const Settings = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = {
    ...getCommonStyles(theme),
    card: {
      borderRadius: '16px',
      background: theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'rgba(255, 255, 255, 0.9)',
      backdropFilter: 'blur(10px)',
      boxShadow: theme.palette.mode === 'dark' 
        ? '0 4px 6px rgba(0, 0, 0, 0.1)' 
        : '0 4px 6px rgba(0, 0, 0, 0.05)',
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
    },
    cardContent: {
      p: 1.5,
      '&:last-child': { pb: 1.5 },
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
    },
    settingsGroup: {
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      minHeight: '200px',
    },
    dnsSettingsGroup: {
      flex: 1,
      display: 'flex',
      flexDirection: 'column',
      minHeight: '140px',
    },
    buttonContainer: {
      mt: 'auto',
      pt: 2,
      display: 'flex',
      justifyContent: 'flex-end',
    },
    input: {
      '& .MuiOutlinedInput-root': {
        borderRadius: '12px',
      },
    },
    select: {
      '& .MuiOutlinedInput-root': {
        borderRadius: '12px',
      },
    },
  };
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [openPasswordDialog, setOpenPasswordDialog] = useState(false);
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  });
  const [remoteAccess, setRemoteAccess] = useState({
    enabled: false,
    domain: '',
    port: '',
    current_domain: '',
    current_port: '',
  });
  const [dnsSettings, setDnsSettings] = useState({
    domestic: '',
    current_domestic: '',
    domestic_type: 'udp',
    singbox_dns: '',
    current_singbox_dns: '',
    singbox_dns_type: 'udp',
    edns_client_subnet: '',
    current_edns_client_subnet: '',
  });
  const [singboxMode, setSingboxMode] = useState({
    mode: 'rule',
  });
  const [fallbackRule, setFallbackRule] = useState('proxy');
  const [currentThemeMode, setCurrentThemeMode] = useState(() => {
    const savedMode = localStorage.getItem('theme_mode') || 'light';
    if (savedMode === 'system') {
      const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
      setMode(isDarkMode ? 'dark' : 'light');
    } else {
      setMode(savedMode);
    }
    return savedMode;
  });
  const [dashboardSettings, setDashboardSettings] = useState({
    type: 'yacd',
    password: '',
    current_password: '',
  });
  const [inboundMode, setInboundMode] = useState('tun');

  // 处理 DNS 设置更改
  const handleDnsSettingsChange = (field, value) => {
    setDnsSettings(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const fetchData = async () => {
    try {
      setLoading(true);
      const settings = await getSettings();
      
      if (settings?.remote_access) {
        setRemoteAccess({
          enabled: settings.remote_access.enabled || false,
          domain: settings.remote_access.domain || '',
          port: settings.remote_access.port || '',
          current_domain: settings.remote_access.current_domain || '',
          current_port: settings.remote_access.current_port || '',
        });
      }

      if (settings?.dns) {
        setDnsSettings({
          domestic: settings.dns.domestic || '',
          current_domestic: settings.dns.current_domestic || '',
          domestic_type: settings.dns.domestic_type || 'udp',
          singbox_dns: settings.dns.singbox_dns || '',
          current_singbox_dns: settings.dns.current_singbox_dns || '',
          singbox_dns_type: settings.dns.singbox_dns_type || 'udp',
          edns_client_subnet: settings.dns.edns_client_subnet || '',
          current_edns_client_subnet: settings.dns.current_edns_client_subnet || '',
        });
      }

      if (settings?.singbox_mode) {
        setSingboxMode({
          mode: settings.singbox_mode.mode || 'rule',
        });
      } else {
        setSingboxMode({
          mode: 'rule',
        });
      }

      if (settings?.inbound_mode) {
        setInboundMode(settings.inbound_mode);
      } else {
        setInboundMode('tun');
      }

      if (settings?.fallback_rule) {
        setFallbackRule(settings.fallback_rule);
      }

      if (settings?.dashboard) {
        setDashboardSettings({
          type: settings.dashboard.type || 'yacd',
          password: '',
          current_password: settings.dashboard.current_password || '',
        });
      }
    } catch (err) {
      console.error('Failed to load settings:', err);
      setError(err.response?.data?.error || '加载设置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleManualRefresh = async () => {
    await fetchData();
  };

  // 预设 DNS 选项
  const presetDns = {
    domestic: {
      udp: [
        { label: '阿里 DNS', value: '223.5.5.5' },
        { label: '腾讯 DNS', value: '119.29.29.29' },
      ],
      tcp: [
        { label: '阿里 DNS', value: '223.5.5.5' },
        { label: '腾讯 DNS', value: '119.29.29.29' },
      ],
      doh: [
        { label: '阿里 DOH', value: 'https://223.5.5.5/dns-query' },
        { label: '阿里 DOH 2', value: 'https://120.53.53.53/dns-query' },
      ],
    },
    singbox: {
      udp: [
        { label: 'Google DNS', value: '8.8.8.8' },
        { label: 'Cloudflare DNS', value: '1.1.1.1' },
      ],
      tcp: [
        { label: 'Google DNS', value: '8.8.8.8' },
        { label: 'Cloudflare DNS', value: '1.1.1.1' },
      ],
      doh: [
        { label: 'Google DOH', value: 'https://dns.google/dns-query' },
        { label: 'Cloudflare DOH', value: 'https://cloudflare-dns.com/dns-query' },
      ],
    },
  };

  // 修改主题模式变更处理函数
  const handleModeChange = (event) => {
    const newMode = event.target.value;
    setCurrentThemeMode(newMode);
    
    // 如果是系统模式，根据系统主题设置
    if (newMode === 'system') {
      const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
      setMode(isDarkMode ? 'dark' : 'light');
    } else {
      setMode(newMode);
    }
    
    // 保存到 localStorage
    localStorage.setItem('theme_mode', newMode);
  };

  // 添加系统主题变化监听
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleThemeChange = (e) => {
      if (currentThemeMode === 'system') {
        setMode(e.matches ? 'dark' : 'light');
      }
    };

    mediaQuery.addEventListener('change', handleThemeChange);

    return () => {
      mediaQuery.removeEventListener('change', handleThemeChange);
    };
  }, [currentThemeMode, setMode]);

  // 处理密码修改
  const handlePasswordChange = async () => {
    // 验证新密码
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      setError('两次输入的新密码不一致');
      return;
    }

    if (passwordForm.newPassword.length < 6) {
      setError('新密码长度不能小于6位');
      return;
    }

    try {
      setLoading(true);
      const response = await updatePassword({
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword,
      });
      
      // 清空表单
      setPasswordForm({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      });
      
      // 显示成功消息
      setSuccess(true);
      setOpenPasswordDialog(false);
      
      // 延迟一秒后执行重新登录操作，让用户看到成功消息
      setTimeout(() => {
        localStorage.removeItem('token');
        window.location.href = '/login';
      }, 1000);
    } catch (err) {
      setError(err.response?.data?.error || '修改密码失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理远程访问设置更新
  const handleRemoteAccessUpdate = async () => {
    try {
      setLoading(true);
      
      // 验证端口号
      const portNum = parseInt(remoteAccess.port);
      if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
        setError('请输入有效的端口号（1-65535）');
        return;
      }

      // 验证域名/IP
      if (!remoteAccess.domain) {
        setError('请输入域名或公网IP');
        return;
      }

      await updateSettings({
        remote_access: {
          enabled: remoteAccess.enabled,
          domain: remoteAccess.domain,
          port: portNum.toString(),
        }
      });

      // 更新当前值显示
      setRemoteAccess(prev => ({
        ...prev,
        current_domain: prev.domain,
        current_port: prev.port,
      }));

      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新远程访问设置失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理 DNS 设置更新
  const handleDnsUpdate = async () => {
    // 添加确认对话框
    if (!window.confirm('确定要保存 DNS 设置吗?')) {
      return;
    }

    try {
      setLoading(true);

      // 验证 DNS 地址格式
      const dnsRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
      const dohRegex = /^https:\/\/.+\/dns-query$/;
      
      // 验证国内 DNS
      if (dnsSettings.domestic) {
        if ((dnsSettings.domestic_type === 'udp' || dnsSettings.domestic_type === 'tcp') && !dnsRegex.test(dnsSettings.domestic)) {
          setError('请输入有效的国内 DNS 地址');
          return;
        }

        if (dnsSettings.domestic_type === 'doh' && !dohRegex.test(dnsSettings.domestic)) {
          setError('请输入有效的国内 DOH 地址');
          return;
        }
      }

      // 验证国外 DNS
      if (dnsSettings.singbox_dns) {
        if ((dnsSettings.singbox_dns_type === 'udp' || dnsSettings.singbox_dns_type === 'tcp') && !dnsRegex.test(dnsSettings.singbox_dns)) {
          setError('请输入有效的国外 DNS 地址');
          return;
        }

        if (dnsSettings.singbox_dns_type === 'doh' && !dohRegex.test(dnsSettings.singbox_dns)) {
          setError('请输入有效的国外 DOH 地址');
          return;
        }
      }

      // 验证 EDNS 客户端子网地址
      if (dnsSettings.edns_client_subnet && !dnsRegex.test(dnsSettings.edns_client_subnet)) {
        setError('请输入有效的 EDNS 客户端子网地址');
        return;
      }

      await updateSettings({
        dns: {
          id: 'default',
          domestic: dnsSettings.domestic,
          domestic_type: dnsSettings.domestic_type,
          singbox_dns: dnsSettings.singbox_dns,
          singbox_dns_type: dnsSettings.singbox_dns_type,
          edns_client_subnet: dnsSettings.edns_client_subnet,
        }
      });

      // 更新当前值显示
      setDnsSettings(prev => ({
        ...prev,
        current_domestic: prev.domestic,
        current_singbox_dns: prev.singbox_dns,
        current_edns_client_subnet: prev.edns_client_subnet,
      }));

      // 保存设置后自动生成配置
      await generateConfig();
      setSuccess(true);
    } catch (err) {
      console.error('Failed to update DNS settings:', err);
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message;
      setError(`更新 DNS 设置失败: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  // 处理漏网之鱼规则更新
  const handleFallbackRuleUpdate = async (value) => {
    try {
      setLoading(true);
      await updateSettings({
        fallback_rule: value
      });
      setFallbackRule(value);
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新漏网之鱼规则失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理 SingBox 模式更新
  const handleSingboxModeUpdate = async (mode) => {
    try {
      setLoading(true);
      await updateSettings({
        singbox_mode: {
          mode: mode,
        }
      });
      setSingboxMode({ mode });
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新 SingBox 模式设置失败');
      // 回滚状态
      setSingboxMode(prev => prev);
    } finally {
      setLoading(false);
    }
  };

  // 更新面板设置
  const updateDashboardSettings = async (settings) => {
    console.log('更新面板设置:', settings);
    const dashboardSettings = {
      type: settings.type,
      password: settings.password || dashboardSettings.current_password,
    };
    console.log('发送到后端的设置:', { dashboard: dashboardSettings });
    await updateSettings({
      dashboard: dashboardSettings
    });
    // 重新生成配置文件
    await generateConfig();
  };

  // 处理面板设置更新
  const handleDashboardUpdate = async () => {
    try {
      setLoading(true);
      setError(null);

      await updateDashboardSettings({
        type: dashboardSettings.type,
        password: dashboardSettings.password,
      });

      setSuccess(true);
      // 更新当前值
      setDashboardSettings(prev => ({
        ...prev,
        current_password: dashboardSettings.password || dashboardSettings.current_password,
        password: '',
      }));
    } catch (err) {
      console.error('Failed to update dashboard settings:', err);
      setError(err.response?.data?.error || '更新仪表盘设置失败');
    } finally {
      setLoading(false);
    }
  };

  // 处理面板类型更改
  const handleDashboardTypeChange = async (value) => {
    try {
      setLoading(true);
      setError(null);
      
      await updateSettings({
        dashboard: {
          type: value,
          secret: dashboardSettings.current_password || ''
        }
      });

      await generateConfig();
      
      setDashboardSettings(prev => ({
        ...prev,
        type: value
      }));
      
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      console.error('Failed to update dashboard type:', err);
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message;
      setError(`更新面板类型失败: ${errorMessage}`);
      // 回滚状态
      setDashboardSettings(prev => prev);
    } finally {
      setLoading(false);
    }
  };

  // 生成配置文件
  const generateConfig = async () => {
    try {
      console.log('开始生成配置文件...');
      setLoading(true);
      const response = await axios.post('/api/config/generate');
      if (response.data?.error) {
        throw new Error(response.data.error);
      }
      console.log('配置文件生成成功');
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      console.error('Failed to generate config:', err);
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message;
      setError(`生成配置失败: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  // 导出设置
  const exportSettings = async () => {
    try {
      setLoading(true);
      const settings = await getSettings();
      const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'singdns-settings.json';
      a.click();
      URL.revokeObjectURL(url);
      setSuccess(true);
    } catch (err) {
      console.error('Failed to export settings:', err);
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message;
      setError(`导出设置失败: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  // 导入设置
  const importSettings = async (file) => {
    try {
      setLoading(true);
      const reader = new FileReader();
      reader.onload = async (e) => {
        try {
          const settings = JSON.parse(e.target.result);
          await updateSettings(settings);
          await generateConfig();
          await fetchData(); // 重新加载设置
          setSuccess(true);
        } catch (err) {
          console.error('Failed to import settings:', err);
          const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message;
          setError(`导入设置失败: ${errorMessage}`);
        } finally {
          setLoading(false);
        }
      };
      reader.readAsText(file);
    } catch (err) {
      console.error('Failed to read settings file:', err);
      setError('读取设置文件失败');
      setLoading(false);
    }
  };

  // 添加文件输入引用
  const fileInputRef = useRef(null);

  // 处理文件选择
  const handleFileSelect = (event) => {
    const file = event.target.files[0];
    if (file) {
      if (file.type !== 'application/json') {
        setError('请选择 JSON 格式的设置文件');
        return;
      }
      importSettings(file);
    }
    // 清除文件输入,以便可以重复选择同一个文件
    event.target.value = '';
  };

  // 处理入站模式更新
  const handleInboundModeUpdate = async (value) => {
    try {
      setLoading(true);
      await updateSettings({
        inbound_mode: value
      });
      setInboundMode(value);
      // 重新生成配置文件
      await generateConfig();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新入站模式设置失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ pb: 2 }}>
      <Snackbar 
        open={success} 
        autoHideDuration={3000} 
        onClose={() => setSuccess(false)}
      >
        <Alert severity="success" sx={{ borderRadius: '12px' }}>操作成功</Alert>
      </Snackbar>

      <Snackbar 
        open={!!error} 
        autoHideDuration={3000} 
        onClose={() => setError(null)}
      >
        <Alert severity="error" sx={{ borderRadius: '12px' }}>{error}</Alert>
      </Snackbar>

      {/* 页面标题卡片 */}
      <Card sx={styles.headerCard}>
        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" alignItems="center" spacing={2}>
              <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
                系统设置
              </Typography>
              {loading && (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
              )}
            </Stack>
            <Stack direction="row" spacing={1}>
              <Button
                size="small"
                variant="outlined"
                onClick={() => fileInputRef.current?.click()}
                disabled={loading}
                sx={{ fontSize: '0.8rem' }}
              >
                导入设置
              </Button>
              <Button
                size="small"
                variant="outlined"
                onClick={exportSettings}
                disabled={loading}
                sx={{ fontSize: '0.8rem' }}
              >
                导出设置
              </Button>
              <IconButton
                size="small"
                onClick={handleManualRefresh}
                disabled={loading}
                sx={styles.iconButton}
                title="刷新"
              >
                <SyncIcon fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={onDashboardClick}
                sx={styles.iconButton}
                title="打开仪表盘"
              >
                <DashboardIcon fontSize="small" />
              </IconButton>
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

      {loading && <LinearProgress sx={{ mb: 1 }} />}

      <Grid container spacing={2} sx={{ px: 2, maxWidth: 1400, margin: '0 auto' }}>
        {/* 第一排：3个卡片 */}
        {/* 基础设置卡片 */}
        <Grid item xs={12} md={4}>
          <Card sx={styles.card}>
            <CardContent sx={styles.cardContent}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 1.5 }}>
                基础设置
              </Typography>
              <Stack spacing={2} sx={styles.settingsGroup}>
                {/* 主题设置 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    主题设置
                  </Typography>
                  <FormControl size="small" fullWidth>
                    <RadioGroup
                      value={currentThemeMode}
                      onChange={handleModeChange}
                    >
                      <Stack direction="row" spacing={2}>
                        <FormControlLabel
                          value="light"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>浅色</span>}
                        />
                        <FormControlLabel
                          value="dark"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>深色</span>}
                        />
                        <FormControlLabel
                          value="system"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>跟随系统</span>}
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>

                {/* 账户安全 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    账户安全
                  </Typography>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<KeyIcon sx={{ fontSize: '0.9rem' }} />}
                    onClick={() => setOpenPasswordDialog(true)}
                    sx={{ fontSize: '0.8rem' }}
                  >
                    修改密码
                  </Button>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 代理模式设置卡片 */}
        <Grid item xs={12} md={4}>
          <Card sx={styles.card}>
            <CardContent sx={styles.cardContent}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 1.5 }}>
                代理模式设置
              </Typography>
              <Stack spacing={2} sx={styles.settingsGroup}>
                {/* 入站模式 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    入站模式
                  </Typography>
                  <FormControl fullWidth>
                    <RadioGroup
                      value={inboundMode}
                      onChange={(e) => handleInboundModeUpdate(e.target.value)}
                    >
                      <Stack direction="row" spacing={2}>
                        <FormControlLabel
                          value="tun"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>TUN 模式</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                系统级透明代理（TCP/UDP）
                              </Typography>
                            </Box>
                          }
                        />
                        <FormControlLabel
                          value="redirect"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>Redirect 模式</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                重定向TCP + Tproxy UDP
                              </Typography>
                            </Box>
                          }
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>

                {/* SingBox 模式 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    SingBox 模式
                  </Typography>
                  <FormControl fullWidth>
                    <RadioGroup
                      value={singboxMode.mode}
                      onChange={(e) => handleSingboxModeUpdate(e.target.value)}
                    >
                      <Stack direction="row" spacing={2}>
                        <FormControlLabel
                          value="rule"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>规则</span>}
                        />
                        <FormControlLabel
                          value="global"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>全局</span>}
                        />
                        <FormControlLabel
                          value="direct"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>直连</span>}
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>

                {/* 漏网之鱼规则 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    漏网之鱼规则
                  </Typography>
                  <FormControl fullWidth>
                    <RadioGroup
                      value={fallbackRule}
                      onChange={(e) => handleFallbackRuleUpdate(e.target.value)}
                    >
                      <Stack direction="row" spacing={2}>
                        <FormControlLabel
                          value="proxy"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>代理</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                未匹配的请求使用代理
                              </Typography>
                            </Box>
                          }
                        />
                        <FormControlLabel
                          value="direct"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>直连</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                未匹配的请求直接连接
                              </Typography>
                            </Box>
                          }
                        />
                        <FormControlLabel
                          value="block"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>拒绝</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                未匹配的请求被拒绝
                              </Typography>
                            </Box>
                          }
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 面板设置卡片 */}
        <Grid item xs={12} md={4}>
          <Card sx={styles.card}>
            <CardContent sx={styles.cardContent}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 1.5 }}>
                面板设置
              </Typography>
              <Stack spacing={2}>
                {/* 面板类型 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    面板类型
                  </Typography>
                  <FormControl fullWidth>
                    <RadioGroup
                      value={dashboardSettings.type}
                      onChange={(e) => handleDashboardTypeChange(e.target.value)}
                    >
                      <Stack direction="row" spacing={2}>
                        <FormControlLabel
                          value="yacd"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>Yacd 面板</span>}
                        />
                        <FormControlLabel
                          value="metacubexd"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>MetaCubeXD 面板</span>}
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>

                {/* 面板访问密码 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    面板访问密码
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <TextField
                      size="small"
                      fullWidth
                      type="password"
                      value={dashboardSettings.password}
                      onChange={(e) => setDashboardSettings(prev => ({
                        ...prev,
                        password: e.target.value
                      }))}
                      placeholder="请输入面板访问密码"
                      helperText={dashboardSettings.current_password ? `当前：${dashboardSettings.current_password}` : ''}
                      FormHelperTextProps={{ sx: { fontSize: '0.7rem' } }}
                      sx={styles.input}
                    />
                  </Stack>
                </Box>
                <Box sx={styles.buttonContainer}>
                  <Button
                    variant="contained"
                    size="small"
                    onClick={handleDashboardUpdate}
                    disabled={loading}
                    sx={{ minWidth: 80 }}
                  >
                    保存
                  </Button>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 第二排：3个卡片 */}
        {/* 国内 DNS 设置卡片 */}
        <Grid item xs={12} md={4}>
          <Card sx={styles.card}>
            <CardContent sx={styles.cardContent}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 1.5 }}>
                国内 DNS 设置
              </Typography>
              <Stack spacing={2} sx={styles.dnsSettingsGroup}>
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    SingBox 国内 DNS（上游服务器）
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <FormControl size="small" fullWidth>
                      {(dnsSettings.domestic_type === 'udp' || dnsSettings.domestic_type === 'tcp' || dnsSettings.domestic_type === 'doh') ? (
                        <Select
                          value={dnsSettings.domestic}
                          onChange={(e) => handleDnsSettingsChange('domestic', e.target.value)}
                          displayEmpty
                          sx={styles.select}
                        >
                          <MenuItem value="" disabled>选择DNS服务器</MenuItem>
                          {presetDns.domestic[dnsSettings.domestic_type].map((preset) => (
                            <MenuItem key={preset.value} value={preset.value}>
                              {preset.label}
                            </MenuItem>
                          ))}
                        </Select>
                      ) : (
                        <TextField
                          size="small"
                          fullWidth
                          value={dnsSettings.domestic}
                          onChange={(e) => handleDnsSettingsChange('domestic', e.target.value)}
                          placeholder="请输入DNS服务器地址"
                          sx={styles.input}
                        />
                      )}
                    </FormControl>
                    <FormControl size="small" sx={{ minWidth: 90 }}>
                      <Select
                        value={dnsSettings.domestic_type}
                        onChange={(e) => handleDnsSettingsChange('domestic_type', e.target.value)}
                        sx={styles.select}
                      >
                        <MenuItem value="udp">UDP</MenuItem>
                        <MenuItem value="tcp">TCP</MenuItem>
                        <MenuItem value="doh">DoH</MenuItem>
                      </Select>
                    </FormControl>
                  </Stack>
                </Box>
                <Box sx={styles.buttonContainer}>
                  <Button
                    variant="contained"
                    size="small"
                    onClick={handleDnsUpdate}
                    disabled={loading}
                    sx={{ minWidth: 80 }}
                  >
                    保存
                  </Button>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 国外 DNS 设置卡片 */}
        <Grid item xs={12} md={4}>
          <Card sx={styles.card}>
            <CardContent sx={styles.cardContent}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 1.5 }}>
                国外 DNS 设置
              </Typography>
              <Stack spacing={2} sx={styles.dnsSettingsGroup}>
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    SingBox 国外 DNS（上游服务器）
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <FormControl size="small" fullWidth>
                      {(dnsSettings.singbox_dns_type === 'udp' || dnsSettings.singbox_dns_type === 'tcp' || dnsSettings.singbox_dns_type === 'doh') ? (
                        <Select
                          value={dnsSettings.singbox_dns}
                          onChange={(e) => handleDnsSettingsChange('singbox_dns', e.target.value)}
                          displayEmpty
                          sx={styles.select}
                        >
                          <MenuItem value="" disabled>选择DNS服务器</MenuItem>
                          {presetDns.singbox[dnsSettings.singbox_dns_type].map((preset) => (
                            <MenuItem key={preset.value} value={preset.value}>
                              {preset.label}
                            </MenuItem>
                          ))}
                        </Select>
                      ) : (
                        <TextField
                          size="small"
                          fullWidth
                          value={dnsSettings.singbox_dns}
                          onChange={(e) => handleDnsSettingsChange('singbox_dns', e.target.value)}
                          placeholder="请输入DNS服务器地址"
                          sx={styles.input}
                        />
                      )}
                    </FormControl>
                    <FormControl size="small" sx={{ minWidth: 90 }}>
                      <Select
                        value={dnsSettings.singbox_dns_type}
                        onChange={(e) => handleDnsSettingsChange('singbox_dns_type', e.target.value)}
                        sx={styles.select}
                      >
                        <MenuItem value="udp">UDP</MenuItem>
                        <MenuItem value="tcp">TCP</MenuItem>
                        <MenuItem value="doh">DoH</MenuItem>
                      </Select>
                    </FormControl>
                  </Stack>
                </Box>
                <Box sx={styles.buttonContainer}>
                  <Button
                    variant="contained"
                    size="small"
                    onClick={handleDnsUpdate}
                    disabled={loading}
                    sx={{ minWidth: 80 }}
                  >
                    保存
                  </Button>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* EDNS 设置卡片 */}
        <Grid item xs={12} md={4}>
          <Card sx={styles.card}>
            <CardContent sx={styles.cardContent}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 1.5 }}>
                EDNS 设置
              </Typography>
              <Stack spacing={2} sx={styles.dnsSettingsGroup}>
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    EDNS 客户端子网
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <TextField
                      size="small"
                      fullWidth
                      value={dnsSettings.edns_client_subnet}
                      onChange={(e) => handleDnsSettingsChange('edns_client_subnet', e.target.value)}
                      placeholder="请输入 IP 地址"
                      helperText={dnsSettings.current_edns_client_subnet ? `当前：${dnsSettings.current_edns_client_subnet}` : ''}
                      FormHelperTextProps={{ sx: { fontSize: '0.7rem' } }}
                      sx={styles.input}
                    />
                  </Stack>
                  <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem', mt: 0.5 }}>
                    用于告知 DNS 服务器客户端的位置，以获得更准确的解析结果
                  </Typography>
                </Box>
                <Box sx={styles.buttonContainer}>
                  <Button
                    variant="contained"
                    size="small"
                    onClick={handleDnsUpdate}
                    disabled={loading}
                    sx={{ minWidth: 80 }}
                  >
                    保存
                  </Button>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* 修改密码对话框 */}
      <Dialog
        open={openPasswordDialog}
        onClose={() => setOpenPasswordDialog(false)}
        maxWidth="xs"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: '16px',
          }
        }}
      >
        <DialogTitle>修改密码</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <TextField
              fullWidth
              type="password"
              label="当前密码"
              value={passwordForm.currentPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, currentPassword: e.target.value })}
              margin="normal"
              size="small"
            />
            <TextField
              fullWidth
              type="password"
              label="新密码"
              value={passwordForm.newPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
              margin="normal"
              size="small"
              helperText="密码长度不能小于6位"
            />
            <TextField
              fullWidth
              type="password"
              label="确认新密码"
              value={passwordForm.confirmPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
              margin="normal"
              size="small"
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenPasswordDialog(false)}>取消</Button>
          <Button onClick={handlePasswordChange} variant="contained">
            确认修改
          </Button>
        </DialogActions>
      </Dialog>

      {/* 添加隐藏的文件输入 */}
      <input
        type="file"
        ref={fileInputRef}
        style={{ display: 'none' }}
        accept=".json"
        onChange={handleFileSelect}
      />
    </Box>
  );
};

export default Settings; 