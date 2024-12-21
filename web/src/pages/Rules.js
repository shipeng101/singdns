import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Button,
  Typography,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Stack,
  Alert,
  Snackbar,
  CircularProgress,
  LinearProgress,
  useTheme,
  FormControlLabel,
  Switch,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  FormHelperText,
  Chip
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon
} from '@mui/icons-material';
import { getRules, createRule, updateRule, deleteRule, getNodes, getNodeGroups } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Rules = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [rules, setRules] = useState([]);
  const [nodes, setNodes] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);
  const [editRule, setEditRule] = useState({
    name: '',
    type: 'remote',
    url: '',
    rules: [],
    outbound: 'direct',
    nodeGroup: '',
    specificNode: '',
    active: true,
  });

  const outboundTypes = [
    { value: 'direct', label: '直连', description: '不走代理直接连接' },
    { value: 'reject', label: '拒绝', description: '通过DNS拒绝连接' },
    { value: 'node_group', label: '节点组', description: '使用指定节点组' },
    { value: 'specific', label: '指定节点', description: '使用特定节点' },
  ];

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [rulesData, nodesData, groupsData] = await Promise.all([
        getRules(),
        getNodes(),
        getNodeGroups()
      ]);
      setRules(rulesData);
      setNodes(nodesData);
      setNodeGroups(groupsData);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.error || '获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (rule = null) => {
    if (rule) {
      setSelectedRule(rule);
      setEditRule({
        ...rule,
        enabled: rule.enabled ?? true
      });
    } else {
      setSelectedRule(null);
      setEditRule({
        name: '',
        type: 'remote',
        url: '',
        rules: [],
        outbound: 'direct',
        nodeGroup: '',
        specificNode: '',
        active: true,
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedRule(null);
    setEditRule({
      name: '',
      type: 'remote',
      url: '',
      rules: [],
      outbound: 'direct',
      nodeGroup: '',
      specificNode: '',
      active: true,
    });
  };

  const handleSaveRule = async () => {
    try {
      setLoading(true);
      const ruleData = {
        ...editRule,
        rules: editRule.type === 'custom' ? editRule.rules : [],
        url: editRule.type === 'remote' ? editRule.url : '',
      };

      if (selectedRule) {
        await updateRule(selectedRule.id, ruleData);
      } else {
        await createRule(ruleData);
      }
      await fetchData();
      setSuccess(true);
      handleCloseDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存规则失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      setLoading(true);
      await deleteRule(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除规则失败');
    } finally {
      setLoading(false);
    }
  };

  const cardVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { 
      opacity: 1, 
      y: 0,
      transition: {
        duration: 0.3,
      }
    }
  };

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1
      }
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
                size="small"
                onClick={onDashboardClick}
                sx={styles.iconButton}
              >
                <OpenInNewIcon fontSize="small" />
              </IconButton>
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

      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">规则列表</Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
          sx={styles.actionButton}
        >
          添加规则
        </Button>
      </Box>

      <Grid container spacing={1.5}>
        {rules.map((rule) => (
          <Grid item xs={12} sm={6} md={4} key={rule.id}>
            <Card sx={styles.card}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip 
                      label={rule.type} 
                      size="small"
                      sx={{
                        ...styles.chip.primary,
                        background: (theme) => 
                          rule.type === 'domain' 
                            ? 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)'
                            : rule.type === 'ip'
                              ? 'linear-gradient(45deg, #42a5f5 30%, #64b5f6 90%)'
                              : 'linear-gradient(45deg, #66bb6a 30%, #81c784 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                    <Typography variant="h6">{rule.name}</Typography>
                  </Stack>
                  <Stack direction="row" spacing={0.5}>
                    <IconButton 
                      size="small" 
                      onClick={() => handleOpenDialog(rule)}
                      sx={{
                        color: 'text.secondary',
                        '&:hover': {
                          color: 'primary.main',
                          backgroundColor: 'action.hover',
                        },
                      }}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton 
                      size="small" 
                      onClick={() => handleDelete(rule.id)}
                      sx={{
                        color: 'text.secondary',
                        '&:hover': {
                          color: 'error.main',
                          backgroundColor: 'action.hover',
                        },
                      }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                </Box>
                <Stack spacing={1}>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      匹配规则
                    </Typography>
                    <Typography variant="body1">
                      {rule.pattern}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      目标节点
                    </Typography>
                    <Typography variant="body1">
                      {rule.target}
                    </Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Chip 
                      label={rule.enabled ? "已启用" : "已禁用"} 
                      size="small"
                      sx={rule.enabled ? {
                        ...styles.chip.success,
                        background: 'linear-gradient(45deg, #4caf50 30%, #81c784 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      } : {
                        ...styles.chip.error,
                        background: 'linear-gradient(45deg, #f44336 30%, #e57373 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                  </Box>
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Dialog 
        open={openDialog} 
        onClose={handleCloseDialog} 
        maxWidth="sm" 
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedRule ? '编辑规则' : '添加规则'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="规则名称"
                  value={editRule.name}
                  onChange={(e) => setEditRule({ ...editRule, name: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>

              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>规则类型</InputLabel>
                  <Select
                    value={editRule.type}
                    label="规则类型"
                    onChange={(e) => setEditRule({ ...editRule, type: e.target.value })}
                    sx={styles.textField}
                  >
                    <MenuItem value="remote">远程规则集</MenuItem>
                    <MenuItem value="custom">自定义规则</MenuItem>
                  </Select>
                  <FormHelperText>
                    {editRule.type === 'remote' ? '从URL导入规则集' : '手动添加规则'}
                  </FormHelperText>
                </FormControl>
              </Grid>

              {editRule.type === 'remote' ? (
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="规则集URL"
                    value={editRule.url}
                    onChange={(e) => setEditRule({ ...editRule, url: e.target.value })}
                    size="small"
                    placeholder="https://example.com/rules.txt"
                    sx={styles.textField}
                  />
                </Grid>
              ) : (
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    rows={6}
                    label="规则列表"
                    value={editRule.rules.join('\n')}
                    onChange={(e) => setEditRule({
                      ...editRule,
                      rules: e.target.value.split('\n').filter(r => r.trim())
                    })}
                    size="small"
                    placeholder="每行一条规则，例如：
DOMAIN-SUFFIX,google.com
DOMAIN-KEYWORD,google
IP-CIDR,8.8.8.8/32"
                    sx={styles.textField}
                  />
                </Grid>
              )}

              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>出口选择</InputLabel>
                  <Select
                    value={editRule.outbound}
                    label="出口选择"
                    onChange={(e) => setEditRule({ ...editRule, outbound: e.target.value })}
                    sx={styles.textField}
                  >
                    {outboundTypes.map((type) => (
                      <MenuItem key={type.value} value={type.value}>
                        <Stack>
                          <Typography variant="body1">{type.label}</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {type.description}
                          </Typography>
                        </Stack>
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>

              {editRule.outbound === 'node_group' && (
                <Grid item xs={12}>
                  <FormControl fullWidth size="small">
                    <InputLabel>节点组</InputLabel>
                    <Select
                      value={editRule.nodeGroup}
                      label="节点组"
                      onChange={(e) => setEditRule({ ...editRule, nodeGroup: e.target.value })}
                      sx={styles.textField}
                    >
                      <MenuItem value="">所有节点组</MenuItem>
                      {nodeGroups.map((group) => (
                        <MenuItem key={group.id} value={group.id}>{group.name}</MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
              )}

              {editRule.outbound === 'specific' && (
                <Grid item xs={12}>
                  <FormControl fullWidth size="small">
                    <InputLabel>指定节点</InputLabel>
                    <Select
                      value={editRule.specificNode}
                      label="指定节点"
                      onChange={(e) => setEditRule({ ...editRule, specificNode: e.target.value })}
                      sx={styles.textField}
                    >
                      {nodes.map((node) => (
                        <MenuItem key={node.id} value={node.id}>{node.name}</MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Grid>
              )}

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editRule.active}
                      onChange={(e) => setEditRule({ ...editRule, active: e.target.checked })}
                    />
                  }
                  label="启用规则"
                />
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={handleCloseDialog} 
            size="small"
            sx={{
              color: 'text.secondary',
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            取消
          </Button>
          <Button 
            variant="contained" 
            onClick={handleSaveRule}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Rules; 