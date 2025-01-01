import React, { useState, useEffect } from 'react';
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
import { getSettings, updateSettings, updatePassword } from '../services/api';

const Settings = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
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
  });
  const [singboxMode, setSingboxMode] = useState({
    mode: 'tproxy',
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
        });
      }

      if (settings?.singbox_mode) {
        setSingboxMode({
          mode: settings.singbox_mode.mode || 'tproxy',
        });
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
    domestic: [
      { label: '阿里 DNS', value: '223.5.5.5', type: 'udp' },
      { label: '腾讯 DNS', value: '119.29.29.29', type: 'udp' },
      { label: '阿里 DOH', value: 'https://dns.alidns.com/dns-query', type: 'doh' },
      { label: '腾讯 DOH', value: 'https://doh.pub/dns-query', type: 'doh' },
    ],
    singbox: [
      { label: 'Google DNS', value: '8.8.8.8', type: 'udp' },
      { label: 'Cloudflare DNS', value: '1.1.1.1', type: 'udp' },
      { label: 'Google DOH', value: 'https://dns.google/dns-query', type: 'doh' },
      { label: 'Cloudflare DOH', value: 'https://cloudflare-dns.com/dns-query', type: 'doh' },
    ],
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
      await updatePassword({
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword,
      });
      setSuccess(true);
      setOpenPasswordDialog(false);
      // 清空表单
      setPasswordForm({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      });
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
    try {
      setLoading(true);

      // 验证 DNS 地址格式
      const dnsRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
      const dohRegex = /^https:\/\/.+\/dns-query$/;
      
      if (dnsSettings.domestic && dnsSettings.domestic_type === 'udp' && !dnsRegex.test(dnsSettings.domestic)) {
        setError('请输入有效的国内 DNS 地址');
        return;
      }

      if (dnsSettings.domestic && dnsSettings.domestic_type === 'doh' && !dohRegex.test(dnsSettings.domestic)) {
        setError('请输入有效的国内 DOH 地址');
        return;
      }

      await updateSettings({
        dns: {
          domestic: dnsSettings.domestic,
          domestic_type: dnsSettings.domestic_type,
          singbox_dns: dnsSettings.singbox_dns,
          singbox_dns_type: dnsSettings.singbox_dns_type,
        }
      });

      // 更新当前值显示
      setDnsSettings(prev => ({
        ...prev,
        current_domestic: prev.domestic,
        current_singbox_dns: prev.singbox_dns,
      }));

      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新 DNS 设置失败');
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

  // 处理面板设置更新
  const handleDashboardUpdate = async () => {
    try {
      setLoading(true);
      await updateSettings({
        dashboard: {
          type: dashboardSettings.type,
          path: dashboardSettings.type === 'metacubexd' ? 'bin/web/metacubexd' : 'bin/web/yacd',
          password: dashboardSettings.password || undefined,
        }
      });
      
      // 更新当前值显示
      setDashboardSettings(prev => ({
        ...prev,
        current_password: prev.password || prev.current_password,
        password: '',
      }));

      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '更新面板设置失败');
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

  return (
    <Box>
      <Snackbar 
        open={success} 
        autoHideDuration={3000} 
        onClose={() => setSuccess(false)}
      >
        <Alert severity="success">操作成功</Alert>
      </Snackbar>

      <Snackbar 
        open={!!error} 
        autoHideDuration={3000} 
        onClose={() => setError(null)}
      >
        <Alert severity="error">{error}</Alert>
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

      <Grid container spacing={1} sx={{ mt: 0 }}>
        {/* 第一列：基础设置和面板设置 */}
        <Grid item xs={12} md={4} sx={{ pr: { md: 0.5 } }}>
          {/* 基础设置卡片 */}
          <Card sx={{ ...styles.card, mb: 1, height: 'calc(50% - 4px)' }}>
            <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 }, height: '100%' }}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 2 }}>
                基础设置
              </Typography>
              
              <Stack spacing={3}>
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

                {/* SingBox 模式设置 */}
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
                          value="tproxy"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>TProxy 模式</span>}
                        />
                        <FormControlLabel
                          value="tun"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>Tun 模式</span>}
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>
              </Stack>
            </CardContent>
          </Card>

          {/* 面板设置卡片 */}
          <Card sx={{ ...styles.card, height: 'calc(50% - 4px)' }}>
            <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 }, height: '100%' }}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 2 }}>
                面板设置
              </Typography>
              
              <Stack spacing={3}>
                {/* 面板类型 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    面板类型
                  </Typography>
                  <FormControl fullWidth size="small">
                    <RadioGroup
                      value={dashboardSettings.type}
                      onChange={(e) => setDashboardSettings(prev => ({ ...prev, type: e.target.value }))}
                    >
                      <Stack direction="row" spacing={2}>
                        <FormControlLabel
                          value="yacd"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>YaCD</span>}
                        />
                        <FormControlLabel
                          value="metacubexd"
                          control={<Radio size="small" />}
                          label={<span style={{ fontSize: '0.8rem' }}>MetaCubeXD</span>}
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>

                {/* 面板密码 */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    面板密码
                  </Typography>
                  <Stack spacing={1}>
                    <TextField
                      fullWidth
                      type="password"
                      value={dashboardSettings.password}
                      onChange={(e) => setDashboardSettings(prev => ({ ...prev, password: e.target.value }))}
                      size="small"
                      placeholder="留空则不修改"
                      helperText={dashboardSettings.current_password ? "当前已设置密码" : "当前未设置密码"}
                      sx={{ '& .MuiInputLabel-root': { fontSize: '0.8rem' } }}
                      InputProps={{ style: { fontSize: '0.8rem' } }}
                      FormHelperTextProps={{ style: { fontSize: '0.7rem' } }}
                    />
                    <Button
                      variant="contained"
                      size="small"
                      onClick={handleDashboardUpdate}
                      disabled={loading}
                      sx={{ fontSize: '0.8rem' }}
                    >
                      保存设置
                    </Button>
                  </Stack>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 第二列：漏网之鱼规则和远程访问设置 */}
        <Grid item xs={12} md={4} sx={{ px: { md: 0.5 } }}>
          {/* 漏网之鱼规则卡片 */}
          <Card sx={{ ...styles.card, mb: 1, height: 'calc(50% - 4px)' }}>
            <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 }, height: '100%' }}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 2 }}>
                漏网之鱼规则
              </Typography>
              
              <Stack spacing={3}>
                <Box>
                  <FormControl fullWidth>
                    <RadioGroup
                      value={fallbackRule}
                      onChange={(e) => handleFallbackRuleUpdate(e.target.value)}
                    >
                      <Stack spacing={2}>
                        <FormControlLabel
                          value="proxy"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>代理</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                未被任何规则匹配的请求将使用代理节点
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
                                未被任何规则匹配的请求将直接连接
                              </Typography>
                            </Box>
                          }
                        />
                        <FormControlLabel
                          value="reject"
                          control={<Radio size="small" />}
                          label={
                            <Box>
                              <Typography sx={{ fontSize: '0.8rem' }}>拒绝</Typography>
                              <Typography sx={{ fontSize: '0.7rem', color: 'text.secondary' }}>
                                未被任何规则匹配的请求将被拒绝
                              </Typography>
                            </Box>
                          }
                        />
                      </Stack>
                    </RadioGroup>
                  </FormControl>
                </Box>

                <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.7rem' }}>
                  设置没有被任何规则命中的请求处理方式
                </Typography>
              </Stack>
            </CardContent>
          </Card>

          {/* 远程访问设置卡片 */}
          <Card sx={{ ...styles.card, height: 'calc(50% - 4px)' }}>
            <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 }, height: '100%' }}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 2 }}>
                远程访问设置
              </Typography>
              
              <Stack spacing={3}>
                <Box>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={remoteAccess.enabled}
                        onChange={(e) => setRemoteAccess(prev => ({ ...prev, enabled: e.target.checked }))}
                        size="small"
                      />
                    }
                    label={<span style={{ fontSize: '0.8rem' }}>启用远程访问</span>}
                  />
                </Box>

                <Stack spacing={2}>
                  <TextField
                    fullWidth
                    label="域名/IP"
                    value={remoteAccess.domain}
                    onChange={(e) => setRemoteAccess(prev => ({ ...prev, domain: e.target.value }))}
                    disabled={!remoteAccess.enabled}
                    size="small"
                    placeholder="example.com"
                    helperText={remoteAccess.current_domain ? `当前：${remoteAccess.current_domain}` : ''}
                    sx={{ '& .MuiInputLabel-root': { fontSize: '0.8rem' } }}
                    InputProps={{ style: { fontSize: '0.8rem' } }}
                    FormHelperTextProps={{ style: { fontSize: '0.7rem' } }}
                  />

                  <TextField
                    fullWidth
                    label="端口"
                    value={remoteAccess.port}
                    onChange={(e) => setRemoteAccess(prev => ({ ...prev, port: e.target.value }))}
                    disabled={!remoteAccess.enabled}
                    size="small"
                    placeholder="443"
                    helperText={remoteAccess.current_port ? `当前：${remoteAccess.current_port}` : ''}
                    sx={{ '& .MuiInputLabel-root': { fontSize: '0.8rem' } }}
                    InputProps={{ style: { fontSize: '0.8rem' } }}
                    FormHelperTextProps={{ style: { fontSize: '0.7rem' } }}
                  />

                  <Button
                    variant="contained"
                    size="small"
                    onClick={handleRemoteAccessUpdate}
                    disabled={!remoteAccess.enabled || loading}
                    sx={{ fontSize: '0.8rem' }}
                  >
                    保存设置
                  </Button>
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 第三列：DNS 设置 */}
        <Grid item xs={12} md={4} sx={{ pl: { md: 0.5 } }}>
          <Card sx={{ ...styles.card, height: '100%' }}>
            <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 }, height: '100%' }}>
              <Typography variant="h6" sx={{ fontSize: '0.9rem', mb: 2 }}>
                DNS 设置
              </Typography>
              
              <Stack spacing={3}>
                {/* 国内 DNS */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    国内 DNS（MosDNS 国内解析服务器）
                  </Typography>
                  <Stack spacing={2}>
                    <FormControl fullWidth size="small">
                      <Select
                        value={dnsSettings.domestic_type}
                        onChange={(e) => setDnsSettings(prev => ({
                          ...prev,
                          domestic_type: e.target.value,
                          domestic: '',
                        }))}
                        sx={{ fontSize: '0.8rem' }}
                      >
                        <MenuItem value="udp" sx={{ fontSize: '0.8rem' }}>UDP DNS</MenuItem>
                        <MenuItem value="doh" sx={{ fontSize: '0.8rem' }}>DOH DNS</MenuItem>
                      </Select>
                    </FormControl>

                    <FormControl fullWidth size="small">
                      <Select
                        value={dnsSettings.domestic}
                        onChange={(e) => setDnsSettings(prev => ({
                          ...prev,
                          domestic: e.target.value,
                        }))}
                        displayEmpty
                        sx={{ fontSize: '0.8rem' }}
                      >
                        <MenuItem value="" sx={{ fontSize: '0.8rem' }}>
                          <em>选择或输入 DNS</em>
                        </MenuItem>
                        {presetDns.domestic
                          .filter(dns => dns.type === dnsSettings.domestic_type)
                          .map(dns => (
                            <MenuItem key={dns.value} value={dns.value} sx={{ fontSize: '0.8rem' }}>
                              {dns.label} - {dns.value}
                            </MenuItem>
                          ))
                        }
                      </Select>
                    </FormControl>

                    <TextField
                      fullWidth
                      value={dnsSettings.domestic}
                      onChange={(e) => setDnsSettings(prev => ({
                        ...prev,
                        domestic: e.target.value
                      }))}
                      size="small"
                      placeholder={dnsSettings.domestic_type === 'udp' ? '223.5.5.5' : 'https://dns.alidns.com/dns-query'}
                      helperText={dnsSettings.current_domestic ? `当前：${dnsSettings.current_domestic}` : ''}
                      sx={{ '& .MuiInputLabel-root': { fontSize: '0.8rem' } }}
                      InputProps={{ style: { fontSize: '0.8rem' } }}
                      FormHelperTextProps={{ style: { fontSize: '0.7rem' } }}
                    />
                  </Stack>
                </Box>

                {/* 国外 DNS */}
                <Box>
                  <Typography variant="subtitle2" sx={{ fontSize: '0.8rem', mb: 1 }}>
                    SingBox 国外 DNS（上游服务器）
                  </Typography>
                  <Stack spacing={2}>
                    <FormControl fullWidth size="small">
                      <Select
                        value={dnsSettings.singbox_dns_type}
                        onChange={(e) => setDnsSettings(prev => ({
                          ...prev,
                          singbox_dns_type: e.target.value,
                          singbox_dns: '',
                        }))}
                        sx={{ fontSize: '0.8rem' }}
                      >
                        <MenuItem value="udp" sx={{ fontSize: '0.8rem' }}>UDP DNS</MenuItem>
                        <MenuItem value="doh" sx={{ fontSize: '0.8rem' }}>DOH DNS</MenuItem>
                      </Select>
                    </FormControl>

                    <FormControl fullWidth size="small">
                      <Select
                        value={dnsSettings.singbox_dns}
                        onChange={(e) => setDnsSettings(prev => ({
                          ...prev,
                          singbox_dns: e.target.value,
                        }))}
                        displayEmpty
                        sx={{ fontSize: '0.8rem' }}
                      >
                        <MenuItem value="" sx={{ fontSize: '0.8rem' }}>
                          <em>选择或输入 DNS</em>
                        </MenuItem>
                        {presetDns.singbox
                          .filter(dns => dns.type === dnsSettings.singbox_dns_type)
                          .map(dns => (
                            <MenuItem key={dns.value} value={dns.value} sx={{ fontSize: '0.8rem' }}>
                              {dns.label} - {dns.value}
                            </MenuItem>
                          ))
                        }
                      </Select>
                    </FormControl>

                    <TextField
                      fullWidth
                      value={dnsSettings.singbox_dns}
                      onChange={(e) => setDnsSettings(prev => ({
                        ...prev,
                        singbox_dns: e.target.value
                      }))}
                      size="small"
                      placeholder={dnsSettings.singbox_dns_type === 'udp' ? '8.8.8.8' : 'https://dns.google/dns-query'}
                      helperText={dnsSettings.current_singbox_dns ? `当前：${dnsSettings.current_singbox_dns}` : ''}
                      sx={{ '& .MuiInputLabel-root': { fontSize: '0.8rem' } }}
                      InputProps={{ style: { fontSize: '0.8rem' } }}
                      FormHelperTextProps={{ style: { fontSize: '0.7rem' } }}
                    />
                  </Stack>
                </Box>

                <Button
                  variant="contained"
                  size="small"
                  onClick={handleDnsUpdate}
                  disabled={loading}
                  sx={{ fontSize: '0.8rem' }}
                >
                  保存设置
                </Button>
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
    </Box>
  );
};

export default Settings; 