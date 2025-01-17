import React, { useState, useEffect } from 'react';
import { useTheme } from '@mui/material/styles';
import {
  Box,
  Card,
  CardContent,
  Typography,
  IconButton,
  CircularProgress,
  LinearProgress,
  Button,
  Stack,
  Grid,
  Chip,
  Snackbar,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Divider,
  Switch,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Sync as SyncIcon,
  LightMode as LightModeIcon,
  DarkMode as DarkModeIcon,
  Dashboard as DashboardIcon,
  Refresh as RefreshIcon,
  Edit as EditIcon,
} from '@mui/icons-material';
import { getCommonStyles } from '../styles/commonStyles';
import { updateRuleSetRules, deleteRuleSet, createRuleSet, getRuleSets, getNodeGroups, updateRuleSet, generateConfig } from '../services/api';
import { formatDistanceToNow } from 'date-fns';
import { zhCN } from 'date-fns/locale';

const Rules = () => {
  const theme = useTheme();
  const [mode, setMode] = useState(theme.palette.mode);
  const styles = getCommonStyles(theme);
  const [loading, setLoading] = useState(false);
  const [updateLoading, setUpdateLoading] = useState({});
  const [snackbar, setSnackbar] = useState({
    open: false,
    message: '',
    severity: 'success'
  });
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deletingRule, setDeletingRule] = useState(null);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [editingRule, setEditingRule] = useState(null);
  const [newRule, setNewRule] = useState({
    name: '',
    type: 'geosite',
    outbound: '节点选择',
    matchContents: []
  });
  const [rules, setRules] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);

  // 获取规则集列表
  const fetchRules = async () => {
    try {
      setLoading(true);
      const response = await getRuleSets();
      console.log('Fetched rules:', response);
      // 确保每个规则都有 updatedAt 字段
      const rulesWithTime = response.map(rule => ({
        ...rule,
        updatedAt: rule.updated_at || rule.updatedAt // 处理可能的不同字段名
      }));
      console.log('Rules with time:', rulesWithTime);
      setRules(rulesWithTime);
    } catch (error) {
      console.error('Failed to fetch rule sets:', error);
      setSnackbar({
        open: true,
        message: '获取规则集列表失败',
        severity: 'error'
      });
    } finally {
      setLoading(false);
    }
  };

  // 在组件加载时获取规则集列表
  useEffect(() => {
    fetchRules();
  }, []);

  // 获取节点组列表
  const fetchNodeGroups = async () => {
    try {
      const data = await getNodeGroups();
      setNodeGroups(data);
    } catch (error) {
      console.error('Failed to fetch node groups:', error);
      setNodeGroups([]);
    }
  };

  // 在组件加载时获取节点组列表
  useEffect(() => {
    fetchNodeGroups();
  }, []);

  // 处理面板点击
  const handleDashboardClick = () => {
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;
    const dashboardUrl = `${protocol}//${hostname}:9090/ui/`;
    window.open(dashboardUrl, '_blank');
  };

  // 处理更新规则集
  const handleUpdateRule = async (ruleId) => {
    try {
      setUpdateLoading(prev => ({ ...prev, [ruleId]: true }));
      const rule = rules.find(r => r.id === ruleId);
      if (!rule) return;

      await updateRuleSetRules(rule.id);
      
      setSnackbar({
        open: true,
        message: '规则集更新成功',
        severity: 'success'
      });
      
      // 刷新规则集列表
      await fetchRules();
    } catch (error) {
      console.error('Failed to update rule set:', error);
      setSnackbar({
        open: true,
        message: '规则集更新失败',
        severity: 'error'
      });
    } finally {
      setUpdateLoading(prev => ({ ...prev, [ruleId]: false }));
    }
  };

  // 处理删除规则集
  const handleDeleteRule = (ruleId) => {
    const rule = rules.find(r => r.id === ruleId);
    if (!rule) return;
    setDeletingRule(rule);
    setDeleteConfirmOpen(true);
  };

  // 确认删除规则集
  const handleConfirmDelete = async () => {
    if (!deletingRule) return;

    try {
      setLoading(true);
      await deleteRuleSet(deletingRule.id);
      
      setSnackbar({
        open: true,
        message: '规则集删除成功',
        severity: 'success'
      });
      
      // 刷新规则列表
      await fetchRules();
    } catch (error) {
      console.error('Failed to delete rule set:', error);
      setSnackbar({
        open: true,
        message: '规则集删除失败',
        severity: 'error'
      });
    } finally {
      setLoading(false);
      setDeleteConfirmOpen(false);
      setDeletingRule(null);
    }
  };

  // 处理关闭提示
  const handleCloseSnackbar = () => {
    setSnackbar(prev => ({ ...prev, open: false }));
  };

  // 处理编辑规则
  const handleEditRule = (rule) => {
    // 获取同组的所有规则
    const groupName = getRuleName(rule.name);
    const groupRules = rules.filter(r => getRuleName(r.name) === groupName);
    
    // 从规则名称中提取类别
    const category = rule.name.split(':')[1];
    setEditingRule(rule);
    setNewRule({
      name: category || '',
      type: rule.type,
      outbound: rule.outbound,
      matchContents: groupRules.map(r => r.url || '').filter(Boolean)
    });
    setAddDialogOpen(true);
  };

  // 处理添加/编辑规则
  const handleAddRule = async () => {
    try {
      setLoading(true);
      // 从规则名称中提取规则集类型和类别
      const ruleSetType = newRule.type.toLowerCase();
      const ruleSetCategory = newRule.name.toLowerCase();
      const ruleSetId = `${ruleSetType}-${ruleSetCategory}`;
      const ruleSetPath = `configs/sing-box/rules/${ruleSetId}.srs`;

      // 构造规则集数据
      const ruleSetData = {
        id: ruleSetId,
        name: `${ruleSetCategory}`,
        type: ruleSetType,
        format: 'binary',
        outbound: newRule.outbound,
        enabled: editingRule ? editingRule.enabled : true,
        path: ruleSetPath
      };

      // 如果是 geosite 或 geoip 规则集，添加 URL
      if (['geosite', 'geoip'].includes(ruleSetType) && newRule.matchContents.length > 0) {
        ruleSetData.url = newRule.matchContents[0];
      } else if (newRule.type !== 'ip_is_private') {
        // 对于其他类型的规则，添加匹配内容
        ruleSetData.matchContents = newRule.matchContents;
      }

      console.log('Creating rule set with data:', ruleSetData);

      if (editingRule) {
        // 更新规则集
        await updateRuleSet(ruleSetId, ruleSetData);
      } else {
        // 创建规则集
        const response = await createRuleSet(ruleSetData);
        console.log('Response:', response);
      }
      
      // 如果是规则集类型，立即更新规则
      if (['geosite', 'geoip'].includes(ruleSetType)) {
        await updateRuleSetRules(ruleSetId);
      }
      
      // 更新配置文件
      await generateConfig();
      
      setSnackbar({
        open: true,
        message: `规则集${editingRule ? '更新' : '添加'}成功`,
        severity: 'success'
      });
      setAddDialogOpen(false);
      setEditingRule(null);
      setNewRule({
        name: '',
        type: 'geosite',
        outbound: '节点选择',
        matchContents: []
      });
      
      // 刷新规则列表
      await fetchRules();
    } catch (error) {
      console.error('Failed to add/update rule set:', error);
      setSnackbar({
        open: true,
        message: `规则集${editingRule ? '更新' : '添加'}失败: ${error.message}`,
        severity: 'error'
      });
    } finally {
      setLoading(false);
    }
  };

  // 处理关闭对话框
  const handleCloseDialog = () => {
    setAddDialogOpen(false);
    setEditingRule(null);
    setNewRule({
      name: '',
      type: 'geosite',
      outbound: '节点选择',
      matchContents: []
    });
  };

  // 获取规则名称的中文名称
  const getRuleName = (name) => {
    // 处理 type:category 格式
    const colonSplit = name.split(':');
    if (colonSplit.length === 2) {
      const category = colonSplit[1];
      return getDisplayName(category);
    }
    
    // 处理 name-type 格式
    const dashSplit = name.split('-');
    if (dashSplit.length >= 2) {
      // 忽略最后一个部分（类型），取前面的部分作为名称
      const category = dashSplit.slice(0, -1).join('-');
      return getDisplayName(category);
    }

    return getDisplayName(name);
  };

  // 获取显示名称
  const getDisplayName = (category) => {
    switch (category.toLowerCase()) {
      case 'category-games':
        return '游戏';
      case 'google':
        return 'Google';
      case 'telegram':
        return 'Telegram';
      case 'youtube':
        return 'YouTube';
      case 'netflix':
        return 'Netflix';
      case 'disney':
        return 'Disney+';
      case 'spotify':
        return 'Spotify';
      case 'apple':
        return 'Apple';
      case 'microsoft':
        return 'Microsoft';
      case 'category-ads-all':
        return '广告';
      case 'cn':
        return '中国大陆';
      case 'geolocation-!cn':
        return '非中国大陆';
      case 'cloudflare':
        return 'Cloudflare';
      default:
        return category;
    }
  };

  // 获取规则类型的中文名称和标签颜色
  const getRuleTypeInfo = (type) => {
    switch (type.toLowerCase()) {
      case 'domain':
        return {
          name: '域名',
          color: theme.palette.primary.main
        };
      case 'ip':
        return {
          name: 'IP',
          color: theme.palette.success.main
        };
      case 'geoip':
        return {
          name: 'GeoIP',
          color: theme.palette.warning.main
        };
      case 'geosite':
        return {
          name: 'GeoSite',
          color: theme.palette.info.main
        };
      default:
        return {
          name: type,
          color: theme.palette.text.secondary
        };
    }
  };

  // 对规则进行分组
  const groupRules = (rules) => {
    const groups = {};
    rules.forEach(rule => {
      const name = getRuleName(rule.name);
      if (!groups[name]) {
        groups[name] = [];
      }
      groups[name].push(rule);
    });
    return Object.entries(groups);
  };

  // 获取匹配内容的提示文本
  const getMatchContentPlaceholder = (type) => {
    switch (type) {
      case 'geosite':
      case 'geoip':
        return '请输入规则集地址，例如：https://example.com/rules/geosite.srs';
      case 'domain':
        return '请输入域名，例如：example.com';
      case 'ip':
        return '请输入 IP，例如：192.168.1.1';
      case 'protocol':
        return '请输入协议，例如：tcp, udp';
      case 'port':
        return '请输入端口，例如：80, 443';
      default:
        return '请输入匹配内容';
    }
  };

  // 处理添加匹配内容
  const handleAddMatchContent = () => {
    setNewRule(prev => ({
      ...prev,
      matchContents: [...prev.matchContents, '']
    }));
  };

  // 处理删除匹配内容
  const handleDeleteMatchContent = (index) => {
    setNewRule(prev => ({
      ...prev,
      matchContents: prev.matchContents.filter((_, i) => i !== index)
    }));
  };

  // 处理匹配内容变更
  const handleMatchContentChange = (index, value) => {
    setNewRule(prev => {
      const newContents = [...prev.matchContents];
      newContents[index] = value;
      return {
        ...prev,
        matchContents: newContents
      };
    });
  };

  // 获取出口显示名称
  const getOutboundDisplayName = (outbound) => {
    if (!outbound) return '未设置';

    // 内置出口
    switch (outbound) {
      case '节点选择':
        return '🚀 节点选择';
      case 'direct':
        return '🎯 直连';
      case 'block':
        return '❌ 拒绝';
      default:
        return outbound;
    }
  };

  // 处理规则启用状态切换
  const handleToggleRule = async (rule) => {
    try {
      setLoading(true);
      // 更新规则状态
      await updateRuleSet(rule.id, {
        ...rule,
        enabled: !rule.enabled
      });
      
      // 更新配置文件
      await generateConfig();
      
      // 刷新规则列表
      await fetchRules();
      
      setSnackbar({
        open: true,
        message: `规则已${!rule.enabled ? '启用' : '禁用'}，配置已更新`,
        severity: 'success'
      });
    } catch (error) {
      console.error('Failed to toggle rule:', error);
      setSnackbar({
        open: true,
        message: '更新规则状态失败',
        severity: 'error'
      });
    } finally {
      setLoading(false);
    }
  };

  // 格式化更新时间
  const formatUpdateTime = (timestamp) => {
    if (!timestamp) return '未更新';
    try {
      console.log('Formatting timestamp:', timestamp);
      const date = new Date(timestamp);
      console.log('Parsed date:', date);
      if (isNaN(date.getTime())) {
        console.error('Invalid date:', timestamp);
        return '未更新';
      }
      return formatDistanceToNow(date, { 
        addSuffix: true,
        locale: zhCN 
      });
    } catch (error) {
      console.error('Error formatting time:', error);
      return '未更新';
    }
  };

  return (
    <Box>
      {/* 渐变色标签栏 */}
      <Card sx={styles.headerCard}>
        <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Stack direction="row" alignItems="center" spacing={2}>
              <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
                规则管理
              </Typography>
              {loading && (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <CircularProgress size={20} thickness={4} sx={{ color: 'rgba(255,255,255,0.8)' }} />
                </Box>
              )}
            </Stack>
            <Stack direction="row" spacing={1}>
              <IconButton
                onClick={() => window.location.reload()}
                sx={styles.iconButton}
                size="small"
              >
                <SyncIcon fontSize="small" />
              </IconButton>
              <IconButton
                onClick={handleDashboardClick}
                sx={styles.iconButton}
                size="small"
                title="打开 Yacd 面板"
              >
                <DashboardIcon fontSize="small" />
              </IconButton>
              <IconButton
                onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
                sx={styles.iconButton}
                size="small"
              >
                {mode === 'dark' ? <LightModeIcon fontSize="small" /> : <DarkModeIcon fontSize="small" />}
              </IconButton>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
      {loading && <LinearProgress sx={{ mb: 1 }} />}

      <Box sx={{ mt: 2 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
          <Typography variant="h6" sx={{ fontWeight: 500 }}>
            规则列表
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setAddDialogOpen(true)}
            sx={styles.actionButton}
          >
            添加规则
          </Button>
        </Stack>

        <Grid container spacing={2}>
          {groupRules(rules).map(([groupName, groupRules]) => (
            <Grid item xs={12} sm={6} md={4} lg={2.4} key={groupName}>
              <Card sx={styles.card}>
                <CardContent sx={{ p: 2, height: 160 }}>
                  <Stack spacing={1.5} height="100%" justifyContent="space-between">
                    <Box>
                      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                        <Typography variant="subtitle1" sx={{ 
                          fontWeight: 500,
                          whiteSpace: 'nowrap',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis'
                        }}>
                          {groupName}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ 
                          fontWeight: 500,
                          whiteSpace: 'nowrap'
                        }}>
                          ({groupRules.length}条规则)
                        </Typography>
                      </Stack>
                      <Box sx={{ 
                        maxHeight: 70,
                        overflowY: 'auto',
                        '&::-webkit-scrollbar': {
                          width: '4px',
                        },
                        '&::-webkit-scrollbar-track': {
                          background: 'transparent',
                        },
                        '&::-webkit-scrollbar-thumb': {
                          background: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)',
                          borderRadius: '2px',
                        },
                        '&::-webkit-scrollbar-thumb:hover': {
                          background: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.2)',
                        },
                      }}>
                        <Stack spacing={1} sx={{ pb: 0.5 }}>
                          <Stack direction="row" spacing={0.5} sx={{ flexWrap: 'wrap', gap: 0.5 }}>
                            {groupRules.map(rule => (
                              <Chip
                                key={rule.id}
                                label={getRuleTypeInfo(rule.type).name}
                                size="small"
                                sx={{
                                  bgcolor: getRuleTypeInfo(rule.type).color + '20',
                                  color: getRuleTypeInfo(rule.type).color,
                                  fontWeight: 500,
                                  fontSize: '0.7rem',
                                }}
                              />
                            ))}
                          </Stack>
                          <Chip
                            label={getOutboundDisplayName(groupRules[0].outbound)}
                            size="small"
                            sx={{
                              maxWidth: 'fit-content',
                              bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                              color: 'text.secondary',
                              fontWeight: 500,
                              fontSize: '0.7rem',
                            }}
                          />
                        </Stack>
                      </Box>
                    </Box>
                    <Stack direction="row" spacing={0.5} alignItems="center" justifyContent="flex-end">
                      <Switch
                        size="small"
                        checked={groupRules.every(rule => rule.enabled !== false)}
                        onChange={() => {
                          const allEnabled = groupRules.every(rule => rule.enabled !== false);
                          groupRules.forEach(rule => handleToggleRule(rule));
                        }}
                        sx={{
                          '& .MuiSwitch-track': {
                            bgcolor: theme.palette.mode === 'dark' 
                              ? 'rgba(255,255,255,0.1)' 
                              : 'rgba(0,0,0,0.1)'
                          }
                        }}
                      />
                      <IconButton 
                        size="small"
                        onClick={() => handleEditRule(groupRules[0])}
                        sx={{ 
                          p: 1,
                          bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.04)',
                          '&:hover': {
                            bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                          }
                        }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton 
                        size="small" 
                        onClick={() => groupRules.forEach(rule => handleUpdateRule(rule.id))}
                        disabled={groupRules.some(rule => updateLoading[rule.id])}
                        sx={{ 
                          p: 1,
                          bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.04)',
                          '&:hover': {
                            bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                          }
                        }}
                      >
                        {groupRules.some(rule => updateLoading[rule.id]) ? (
                          <CircularProgress size={18} thickness={4} />
                        ) : (
                          <RefreshIcon fontSize="small" />
                        )}
                      </IconButton>
                      <IconButton 
                        size="small" 
                        onClick={() => handleDeleteRule(groupRules[0].id)}
                        sx={{ 
                          p: 1,
                          bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.04)',
                          '&:hover': {
                            bgcolor: theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                          }
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Box>

      {/* 删除确认对话框 */}
      <Dialog
        open={deleteConfirmOpen}
        onClose={() => setDeleteConfirmOpen(false)}
        maxWidth="xs"
        fullWidth
      >
        <DialogTitle>确认删除</DialogTitle>
        <DialogContent>
          <Typography>
            确定要删除规则集 "{deletingRule?.name}" 吗？删除后将重新生成 sing-box 配置文件。
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteConfirmOpen(false)}>取消</Button>
          <Button
            onClick={handleConfirmDelete}
            color="error"
            variant="contained"
            disabled={loading}
          >
            {loading ? <CircularProgress size={24} /> : '删除'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* 添加/编辑规则对话框 */}
      <Dialog
        open={addDialogOpen}
        onClose={handleCloseDialog}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: 3,
            bgcolor: theme.palette.mode === 'dark' 
              ? 'rgba(22,28,36,0.8)'
              : 'rgba(255,255,255,0.8)',
            backdropFilter: 'blur(20px)',
            boxShadow: theme.palette.mode === 'dark'
              ? '0 8px 16px 0 rgba(0,0,0,0.4)'
              : '0 8px 16px 0 rgba(145,158,171,0.16)'
          }
        }}
      >
        <DialogTitle>{editingRule ? '编辑规则' : '添加规则'}</DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <TextField
              label="名称"
              required
              value={newRule.name}
              onChange={(e) => setNewRule({ ...newRule, name: e.target.value })}
              placeholder="请输入规则名称"
              sx={{
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2,
                  bgcolor: theme.palette.mode === 'dark' 
                    ? 'rgba(255,255,255,0.05)'
                    : 'rgba(0,0,0,0.02)',
                  '&:hover': {
                    bgcolor: theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.08)'
                      : 'rgba(0,0,0,0.04)'
                  }
                }
              }}
            />
            <FormControl>
              <InputLabel>类型</InputLabel>
              <Select
                value={newRule.type}
                label="类型"
                onChange={(e) => {
                  setNewRule({ 
                    ...newRule, 
                    type: e.target.value,
                    matchContents: []  // 切换类型时清空匹配内容
                  });
                }}
                sx={{
                  borderRadius: 2,
                  bgcolor: theme.palette.mode === 'dark' 
                    ? 'rgba(255,255,255,0.05)'
                    : 'rgba(0,0,0,0.02)',
                  '&:hover': {
                    bgcolor: theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.08)'
                      : 'rgba(0,0,0,0.04)'
                  }
                }}
              >
                <MenuItem value="geosite">GeoSite 规则集</MenuItem>
                <MenuItem value="geoip">GeoIP 规则集</MenuItem>
                <MenuItem value="domain">域名规则</MenuItem>
                <MenuItem value="ip">IP 规则</MenuItem>
                <MenuItem value="ip_is_private">私有 IP 规则</MenuItem>
                <MenuItem value="protocol">协议规则</MenuItem>
                <MenuItem value="port">端口规则</MenuItem>
              </Select>
            </FormControl>
            <FormControl>
              <InputLabel>出口</InputLabel>
              <Select
                value={newRule.outbound}
                label="出口"
                onChange={(e) => setNewRule({ ...newRule, outbound: e.target.value })}
                sx={{
                  borderRadius: 2,
                  bgcolor: theme.palette.mode === 'dark' 
                    ? 'rgba(255,255,255,0.05)'
                    : 'rgba(0,0,0,0.02)',
                  '&:hover': {
                    bgcolor: theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.08)'
                      : 'rgba(0,0,0,0.04)'
                  }
                }}
              >
                <MenuItem value="节点选择">🚀 节点选择</MenuItem>
                <MenuItem value="direct">🎯 直连</MenuItem>
                <MenuItem value="block">❌ 拒绝</MenuItem>
                {nodeGroups.map(group => (
                  <MenuItem key={group.tag} value={group.tag}>{group.tag}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <Box sx={{ 
              border: 1, 
              borderColor: theme.palette.mode === 'dark'
                ? 'rgba(255,255,255,0.1)'
                : 'rgba(0,0,0,0.1)',
              borderRadius: 2,
              p: 2,
            }}>
              <Stack spacing={2}>
                <Stack direction="row" justifyContent="space-between" alignItems="center">
                  <Typography color="text.secondary">
                    {editingRule ? '规则更新地址' : ['geosite', 'geoip'].includes(newRule.type) ? '规则集地址' : '匹配内容'}
                  </Typography>
                  {!editingRule && (
                    <Button
                      startIcon={<AddIcon />}
                      onClick={handleAddMatchContent}
                      disabled={newRule.type === 'ip_is_private'}
                    >
                      添加
                    </Button>
                  )}
                </Stack>
                {newRule.matchContents.length > 0 ? (
                  <Stack spacing={1}>
                    {newRule.matchContents.map((content, index) => (
                      <Stack
                        key={index}
                        direction="row"
                        spacing={1}
                        alignItems="center"
                      >
                        <TextField
                          fullWidth
                          size="small"
                          required
                          value={content}
                          onChange={(e) => handleMatchContentChange(index, e.target.value)}
                          placeholder={getMatchContentPlaceholder(newRule.type)}
                          disabled={editingRule}
                          sx={{
                            '& .MuiOutlinedInput-root': {
                              borderRadius: 2,
                              bgcolor: theme.palette.mode === 'dark' 
                                ? 'rgba(255,255,255,0.05)'
                                : 'rgba(0,0,0,0.02)',
                              '&:hover': {
                                bgcolor: theme.palette.mode === 'dark'
                                  ? 'rgba(255,255,255,0.08)'
                                  : 'rgba(0,0,0,0.04)'
                              }
                            }
                          }}
                        />
                        {!editingRule && (
                          <IconButton 
                            size="small" 
                            onClick={() => handleDeleteMatchContent(index)}
                            sx={{
                              color: theme.palette.error.main,
                              '&:hover': {
                                bgcolor: theme.palette.error.main + '20'
                              }
                            }}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        )}
                      </Stack>
                    ))}
                  </Stack>
                ) : (
                  <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 2 }}>
                    {editingRule ? '暂无更新地址' : '暂无匹配内容'}
                  </Typography>
                )}
              </Stack>
            </Box>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>取消</Button>
          <Button
            onClick={handleAddRule}
            variant="contained"
            disabled={loading || !newRule.name || (newRule.matchContents.length === 0 && newRule.type !== 'ip_is_private')}
          >
            {loading ? <CircularProgress size={24} /> : editingRule ? '保存' : '添加'}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Rules; 