import { useAppStore } from '@/shared/store/useAppStore';
import { Toast } from './Toast';

export const ToastContainer = () => {
  const { notifications, removeNotification } = useAppStore();

  return (
    <div className="fixed bottom-0 right-0 z-50 m-8 flex flex-col gap-2">
      {notifications.map((notification) => (
        <Toast
          key={notification.id}
          {...notification}
          onClose={removeNotification}
        />
      ))}
    </div>
  );
}; 