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
  Chip,
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
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Divider
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  DarkMode as DarkModeIcon,
  LightMode as LightModeIcon,
  OpenInNew as OpenInNewIcon,
  Sync as SyncIcon,
  CloudSync as CloudSyncIcon
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { getNodes, getNodeGroups, createNodeGroup, updateNodeGroup, deleteNodeGroup, importNodes } from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';

const Nodes = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);
  const [nodes, setNodes] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [openGroupDialog, setOpenGroupDialog] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [editGroup, setEditGroup] = useState({
    name: '',
    tag: '',
    includePatterns: [],
    excludePatterns: [],
    mode: 'select',
    active: true
  });
  const [openNodeDialog, setOpenNodeDialog] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [editNode, setEditNode] = useState({
    name: '',
    type: 'ss',
    server: '',
    port: '',
    method: 'aes-256-gcm',
    password: '',
    uuid: '',
    alterId: 0,
    security: 'auto',
    sni: '',
    skipCertVerify: false,
    protocol: 'udp',
    up: '',
    down: '',
    obfs: '',
    alpn: [],
    uploadMbps: 100,
    downloadMbps: 100,
    version: 3,
    username: '',
    tls: false,
    network: 'tcp',
    udp: true,
    ws: false,
    wsPath: '',
    wsHeaders: {},
  });
  const [openImportDialog, setOpenImportDialog] = useState(false);
  const [importUrl, setImportUrl] = useState('');
  const [autoUpdate, setAutoUpdate] = useState(false);
  const [updateInterval, setUpdateInterval] = useState('');

  const nodeTypes = [
    { value: 'ss', label: 'Shadowsocks' },
    { value: 'vmess', label: 'VMess' },
    { value: 'trojan', label: 'Trojan' },
    { value: 'naive', label: 'Naive' },
    { value: 'hysteria', label: 'Hysteria' },
    { value: 'hysteria2', label: 'Hysteria2' },
    { value: 'shadowtls', label: 'ShadowTLS' },
    { value: 'tun', label: 'Tun' },
    { value: 'redirect', label: 'Redirect' },
    { value: 'tproxy', label: 'TProxy' },
    { value: 'socks', label: 'Socks' },
    { value: 'http', label: 'HTTP' },
  ];

  const ssEncryption = [
    'aes-128-gcm',
    'aes-256-gcm',
    'chacha20-poly1305',
    'xchacha20-poly1305',
    '2022-blake3-aes-128-gcm',
    '2022-blake3-aes-256-gcm',
    '2022-blake3-chacha20-poly1305',
  ];

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [nodesData, groupsData] = await Promise.all([
        getNodes(),
        getNodeGroups()
      ]);
      setNodes(nodesData);
      setNodeGroups(groupsData);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.error || '获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenGroupDialog = (group = null) => {
    if (group) {
      setSelectedGroup(group);
      setEditGroup({
        ...group,
        includePatterns: group.includePatterns || [],
        excludePatterns: group.excludePatterns || [],
        mode: group.mode || 'select',
        tag: group.tag || '',
        active: group.active ?? true
      });
    } else {
      setSelectedGroup(null);
      setEditGroup({
        name: '',
        tag: '',
        includePatterns: [],
        excludePatterns: [],
        mode: 'select',
        active: true
      });
    }
    setOpenGroupDialog(true);
  };

  const handleCloseGroupDialog = () => {
    setOpenGroupDialog(false);
    setSelectedGroup(null);
    setEditGroup({
      name: '',
      tag: '',
      includePatterns: [],
      excludePatterns: [],
      mode: 'select',
      active: true
    });
  };

  const handleSaveGroup = async () => {
    try {
      setLoading(true);
      if (selectedGroup) {
        await updateNodeGroup(selectedGroup.id, editGroup);
      } else {
        await createNodeGroup(editGroup);
      }
      await fetchData();
      setSuccess(true);
      handleCloseGroupDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存分组失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteGroup = async (id) => {
    try {
      setLoading(true);
      await deleteNodeGroup(id);
      await fetchData();
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.error || '删除分组失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenNodeDialog = (node = null) => {
    if (node) {
      setSelectedNode(node);
      setEditNode({
        ...node,
        name: node.name || '',
        type: node.type || 'ss',
        server: node.server || '',
        port: node.port || '',
        method: node.method || 'aes-256-gcm',
        password: node.password || '',
        uuid: node.uuid || '',
        alterId: node.alterId || 0,
        security: node.security || 'auto',
        sni: node.sni || '',
        skipCertVerify: node.skipCertVerify || false,
        protocol: node.protocol || 'udp',
        up: node.up || '',
        down: node.down || '',
        obfs: node.obfs || '',
        alpn: node.alpn || [],
        uploadMbps: node.uploadMbps || 100,
        downloadMbps: node.downloadMbps || 100,
        version: node.version || 3,
        username: node.username || '',
        tls: node.tls || false,
        network: node.network || 'tcp',
        udp: node.udp || true,
        ws: node.ws || false,
        wsPath: node.wsPath || '',
        wsHeaders: node.wsHeaders || {},
      });
    } else {
      setSelectedNode(null);
      setEditNode({
        name: '',
        type: 'ss',
        server: '',
        port: '',
        method: 'aes-256-gcm',
        password: '',
        uuid: '',
        alterId: 0,
        security: 'auto',
        sni: '',
        skipCertVerify: false,
        protocol: 'udp',
        up: '',
        down: '',
        obfs: '',
        alpn: [],
        uploadMbps: 100,
        downloadMbps: 100,
        version: 3,
        username: '',
        tls: false,
        network: 'tcp',
        udp: true,
        ws: false,
        wsPath: '',
        wsHeaders: {},
      });
    }
    setOpenNodeDialog(true);
  };

  const handleCloseNodeDialog = () => {
    setOpenNodeDialog(false);
    setSelectedNode(null);
    setEditNode({
      name: '',
      type: 'ss',
      server: '',
      port: '',
      method: 'aes-256-gcm',
      password: '',
      uuid: '',
      alterId: 0,
      security: 'auto',
      sni: '',
      skipCertVerify: false,
      protocol: 'udp',
      up: '',
      down: '',
      obfs: '',
      alpn: [],
      uploadMbps: 100,
      downloadMbps: 100,
      version: 3,
      username: '',
      tls: false,
      network: 'tcp',
      udp: true,
      ws: false,
      wsPath: '',
      wsHeaders: {},
    });
  };

  const handleSaveNode = async () => {
    try {
      setLoading(true);
      if (selectedNode) {
        await updateNodeGroup(selectedNode.id, editNode);
      } else {
        await createNodeGroup(editNode);
      }
      await fetchData();
      setSuccess(true);
      handleCloseNodeDialog();
    } catch (err) {
      setError(err.response?.data?.error || '保存节点失败');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenImportDialog = () => {
    setImportDialogOpen(true);
  };

  const handleCloseImportDialog = () => {
    setImportDialogOpen(false);
    setImportContent('');
    setImportType('subscription');
  };

  const handleImportNodes = async () => {
    try {
      setLoading(true);
      const nodeLinks = importContent.split('\n').filter(link => link.trim());
      for (const link of nodeLinks) {
        const protocol = link.split('://')[0].toLowerCase();
        if (['ss', 'vmess', 'trojan', 'naive', 'hysteria', 'hysteria2', 'shadowtls', 'tun', 'redirect', 'tproxy', 'socks', 'http'].includes(protocol)) {
          await importNodes({ url: link, type: protocol });
        }
      }
      await fetchData();
      setSuccess(true);
      handleCloseImportDialog();
    } catch (error) {
      setError(error.response?.data?.error || '导入节点失败');
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

  const renderNodeFields = () => {
    switch (editNode.type) {
      case 'ss':
        return (
          <>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>加密方式</InputLabel>
                <Select
                  value={editNode.method}
                  label="加密方式"
                  onChange={(e) => setEditNode({ ...editNode, method: e.target.value })}
                  sx={styles.textField}
                >
                  {ssEncryption.map((method) => (
                    <MenuItem key={method} value={method}>{method}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
          </>
        );

      case 'vmess':
        return (
          <>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="UUID"
                value={editNode.uuid}
                onChange={(e) => setEditNode({ ...editNode, uuid: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="AlterID"
                type="number"
                value={editNode.alterId}
                onChange={(e) => setEditNode({ ...editNode, alterId: parseInt(e.target.value) || 0 })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>加密方式</InputLabel>
                <Select
                  value={editNode.security}
                  label="加密方式"
                  onChange={(e) => setEditNode({ ...editNode, security: e.target.value })}
                  sx={styles.textField}
                >
                  <MenuItem value="auto">auto</MenuItem>
                  <MenuItem value="aes-128-gcm">aes-128-gcm</MenuItem>
                  <MenuItem value="chacha20-poly1305">chacha20-poly1305</MenuItem>
                  <MenuItem value="none">none</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </>
        );

      case 'trojan':
        return (
          <>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="SNI"
                value={editNode.sni}
                onChange={(e) => setEditNode({ ...editNode, sni: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={editNode.skipCertVerify}
                    onChange={(e) => setEditNode({ ...editNode, skipCertVerify: e.target.checked })}
                  />
                }
                label="跳过证书验证"
              />
            </Grid>
          </>
        );

      case 'hysteria':
      case 'hysteria2':
        return (
          <>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="上行速度 (Mbps)"
                type="number"
                value={editNode.type === 'hysteria' ? editNode.up : editNode.uploadMbps}
                onChange={(e) => setEditNode({
                  ...editNode,
                  [editNode.type === 'hysteria' ? 'up' : 'uploadMbps']: e.target.value
                })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="下行速度 (Mbps)"
                type="number"
                value={editNode.type === 'hysteria' ? editNode.down : editNode.downloadMbps}
                onChange={(e) => setEditNode({
                  ...editNode,
                  [editNode.type === 'hysteria' ? 'down' : 'downloadMbps']: e.target.value
                })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            {editNode.type === 'hysteria' && (
              <>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="混淆密码"
                    value={editNode.obfs}
                    onChange={(e) => setEditNode({ ...editNode, obfs: e.target.value })}
                    size="small"
                    sx={styles.textField}
                  />
                </Grid>
                <Grid item xs={12}>
                  <FormControl fullWidth size="small">
                    <InputLabel>ALPN</InputLabel>
                    <Select
                      multiple
                      value={editNode.alpn}
                      label="ALPN"
                      onChange={(e) => setEditNode({ ...editNode, alpn: e.target.value })}
                      sx={styles.textField}
                    >
                      <MenuItem value="h3">h3</MenuItem>
                      <MenuItem value="h2">h2</MenuItem>
                      <MenuItem value="http/1.1">http/1.1</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
              </>
            )}
          </>
        );

      case 'shadowtls':
        return (
          <>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth size="small">
                <InputLabel>版本</InputLabel>
                <Select
                  value={editNode.version}
                  label="版本"
                  onChange={(e) => setEditNode({ ...editNode, version: e.target.value })}
                  sx={styles.textField}
                >
                  <MenuItem value={1}>v1</MenuItem>
                  <MenuItem value={2}>v2</MenuItem>
                  <MenuItem value={3}>v3</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </>
        );

      case 'naive':
        return (
          <>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="用户名"
                value={editNode.username}
                onChange={(e) => setEditNode({ ...editNode, username: e.target.value })}
                size="small"
                sx={styles.textField}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="密码"
                value={editNode.password}
                onChange={(e) => setEditNode({ ...editNode, password: e.target.value })}
                size="small"
                type="password"
                sx={styles.textField}
              />
            </Grid>
          </>
        );

      default:
        return null;
    }
  };

  // 导入节点对话框
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [importType, setImportType] = useState('subscription');
  const [importContent, setImportContent] = useState('');

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
                节点管理
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

      {/* Nodes List */}
      <Box sx={{ mb: 3 }}>
        <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="h6">节点列表</Typography>
          <Stack direction="row" spacing={1}>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={() => handleOpenNodeDialog()}
              sx={styles.actionButton}
            >
              添加节点
            </Button>
            <Button
              variant="contained"
              size="small"
              startIcon={<CloudSyncIcon />}
              onClick={handleOpenImportDialog}
              sx={styles.actionButton}
            >
              导入节点
            </Button>
            <Button
              variant="contained"
              size="small"
              startIcon={<SyncIcon />}
              onClick={fetchData}
              disabled={loading}
              sx={styles.actionButton}
            >
              刷新节点
            </Button>
          </Stack>
        </Box>
        <TableContainer component={Paper} sx={styles.tableContainer}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>名称</TableCell>
                <TableCell>类型</TableCell>
                <TableCell>地址</TableCell>
                <TableCell>延迟</TableCell>
                <TableCell>状态</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {nodes.map((node) => (
                <TableRow key={node.id}>
                  <TableCell>{node.name}</TableCell>
                  <TableCell>{node.type}</TableCell>
                  <TableCell>{node.address}</TableCell>
                  <TableCell>{node.latency}ms</TableCell>
                  <TableCell>
                    <Chip 
                      label={node.status} 
                      size="small"
                      sx={node.status === 'online' ? {
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
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>

      {/* Node Groups */}
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">节点分组</Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={() => handleOpenGroupDialog()}
          sx={styles.actionButton}
        >
          添加分组
        </Button>
      </Box>

      <Grid container spacing={1.5}>
        {nodeGroups.map((group) => (
          <Grid item xs={12} sm={6} md={3} key={group.id}>
            <Card sx={styles.card}>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip 
                      label={group.tag} 
                      size="small"
                      sx={{
                        ...styles.chip.primary,
                        background: 'linear-gradient(45deg, #3949ab 30%, #5e35b1 90%)',
                        color: '#fff',
                        fontWeight: 500,
                      }}
                    />
                    <Typography variant="h6">{group.name}</Typography>
                  </Stack>
                  <Stack direction="row" spacing={0.5}>
                    <IconButton 
                      size="small" 
                      onClick={() => handleOpenGroupDialog(group)}
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
                      onClick={() => handleDeleteGroup(group.id)}
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
                <Stack direction="row" spacing={1} alignItems="center">
                  <Typography variant="body2" color="text.secondary">
                    {group.nodeCount || 0} 个节点
                  </Typography>
                  <Chip 
                    label={group.active ? "活跃" : "未激活"} 
                    size="small"
                    sx={group.active ? {
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
                  <Chip 
                    label={group.mode === 'select' ? "手动选择" : "自动测速"} 
                    size="small"
                    sx={group.mode === 'select' ? {
                      ...styles.chip.warning,
                      background: 'linear-gradient(45deg, #ff9800 30%, #ffb74d 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    } : {
                      ...styles.chip.success,
                      background: 'linear-gradient(45deg, #42a5f5 30%, #64b5f6 90%)',
                      color: '#fff',
                      fontWeight: 500,
                    }}
                  />
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Group Dialog */}
      <Dialog 
        open={openGroupDialog} 
        onClose={handleCloseGroupDialog} 
        maxWidth="sm" 
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedGroup ? '编辑分组' : '添加分组'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="分组名称"
                  value={editGroup.name}
                  onChange={(e) => setEditGroup({ ...editGroup, name: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="分组标签"
                  value={editGroup.tag}
                  onChange={(e) => setEditGroup({ ...editGroup, tag: e.target.value })}
                  size="small"
                  placeholder="例如：HK, JP, SG, US"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>节点选择模式</InputLabel>
                  <Select
                    value={editGroup.mode}
                    label="节点选择模式"
                    onChange={(e) => setEditGroup({ ...editGroup, mode: e.target.value })}
                    sx={styles.textField}
                  >
                    <MenuItem value="select">手动选择</MenuItem>
                    <MenuItem value="urltest">自动测速</MenuItem>
                  </Select>
                  <FormHelperText>
                    {editGroup.mode === 'select' ? '手动选择节点' : '自动选择可用和延迟较低的节点'}
                  </FormHelperText>
                </FormControl>
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editGroup.active}
                      onChange={(e) => setEditGroup({ ...editGroup, active: e.target.checked })}
                    />
                  }
                  label="活跃"
                />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="subtitle2" gutterBottom>
                  节点匹配规则
                </Typography>
                <Stack spacing={2}>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      包含节点（白名单模式）
                    </Typography>
                    <TextField
                      fullWidth
                      multiline
                      rows={3}
                      placeholder="每行一个正则表达式，例如：
香港|HK|🇭🇰
新加坡|SG|🇸🇬
日本|JP|🇯🇵
美国|US|🇺🇸"
                      value={editGroup.includePatterns.join('\n')}
                      onChange={(e) => setEditGroup({
                        ...editGroup,
                        includePatterns: e.target.value.split('\n').filter(p => p.trim())
                      })}
                      size="small"
                      sx={styles.textField}
                    />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary" gutterBottom>
                      排除节点（黑名单模式）
                    </Typography>
                    <TextField
                      fullWidth
                      multiline
                      rows={3}
                      placeholder="每行一个正则表达式，例如：.*过期.*"
                      value={editGroup.excludePatterns.join('\n')}
                      onChange={(e) => setEditGroup({
                        ...editGroup,
                        excludePatterns: e.target.value.split('\n').filter(p => p.trim())
                      })}
                      size="small"
                      sx={styles.textField}
                    />
                  </Box>
                </Stack>
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button 
            onClick={handleCloseGroupDialog} 
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
            onClick={handleSaveGroup}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* Node Dialog */}
      <Dialog
        open={openNodeDialog}
        onClose={handleCloseNodeDialog}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>
          {selectedNode ? '编辑节点' : '添加节点'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="节点名称"
                  value={editNode.name}
                  onChange={(e) => setEditNode({ ...editNode, name: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth size="small">
                  <InputLabel>节点类型</InputLabel>
                  <Select
                    value={editNode.type}
                    label="节点类型"
                    onChange={(e) => setEditNode({ ...editNode, type: e.target.value })}
                    sx={styles.textField}
                  >
                    {nodeTypes.map((type) => (
                      <MenuItem key={type.value} value={type.value}>{type.label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="服务器地址"
                  value={editNode.server}
                  onChange={(e) => setEditNode({ ...editNode, server: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="端口"
                  type="number"
                  value={editNode.port}
                  onChange={(e) => setEditNode({ ...editNode, port: e.target.value })}
                  size="small"
                  sx={styles.textField}
                />
              </Grid>
              {renderNodeFields()}
              <Grid item xs={12}>
                <Divider sx={{ my: 1 }}>高级设置</Divider>
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.tls}
                      onChange={(e) => setEditNode({ ...editNode, tls: e.target.checked })}
                    />
                  }
                  label="启用 TLS"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.udp}
                      onChange={(e) => setEditNode({ ...editNode, udp: e.target.checked })}
                    />
                  }
                  label="启用 UDP"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={editNode.ws}
                      onChange={(e) => setEditNode({ ...editNode, ws: e.target.checked })}
                    />
                  }
                  label="WebSocket"
                />
              </Grid>
              {editNode.ws && (
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="WebSocket 路径"
                    value={editNode.wsPath}
                    onChange={(e) => setEditNode({ ...editNode, wsPath: e.target.value })}
                    size="small"
                    sx={styles.textField}
                  />
                </Grid>
              )}
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={handleCloseNodeDialog}
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
            onClick={handleSaveNode}
            disabled={loading}
            size="small"
            sx={styles.actionButton}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>

      {/* 导入节点对话框 */}
      <Dialog
        open={importDialogOpen}
        onClose={handleCloseImportDialog}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: styles.dialog
        }}
      >
        <DialogTitle>导入节点</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  multiline
                  rows={6}
                  label="节点链接"
                  value={importContent}
                  onChange={(e) => setImportContent(e.target.value)}
                  size="small"
                  placeholder="支持以下格式节点链接，每行一个：
ss://...
vmess://...
trojan://...
naive://...
hysteria://...
hysteria2://...
shadowtls://...
tun://...
redirect://...
tproxy://...
socks://...
http://..."
                  sx={styles.textField}
                />
                <FormHelperText>
                  支持 SS、VMess、Trojan、Naive、Hysteria、Hysteria2、ShadowTLS、Tun、Redirect、TProxy、Socks、HTTP 等格式节点链接，每行一个
                </FormHelperText>
              </Grid>
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={handleCloseImportDialog}
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
            onClick={handleImportNodes}
            disabled={loading || !importContent}
            size="small"
            sx={styles.actionButton}
          >
            导入
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Nodes; 