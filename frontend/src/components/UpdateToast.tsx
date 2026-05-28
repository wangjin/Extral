import { UpdateInfo } from '../hooks/useUpdate';

interface Props {
  info: UpdateInfo;
  onViewDetails: () => void;
  onDismiss: () => void;
}

export default function UpdateToast({ info, onViewDetails, onDismiss }: Props) {
  return (
    <div className="update-toast">
      <div className="update-toast-title">发现新版本 {info.version}</div>
      <div className="update-toast-actions">
        <button className="btn btn-primary" onClick={onViewDetails}>查看详情</button>
        <button className="btn btn-secondary" onClick={onDismiss}>忽略</button>
      </div>
    </div>
  );
}
