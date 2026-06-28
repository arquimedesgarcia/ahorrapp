import 'package:flutter/material.dart';

class ReceiptStatusChip extends StatelessWidget {
  const ReceiptStatusChip({super.key, required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final config = _chipConfig(status);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: config.color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(config.icon, size: 16, color: config.color),
          const SizedBox(width: 4),
          Text(
            config.label,
            style: TextStyle(color: config.color, fontSize: 12),
          ),
        ],
      ),
    );
  }

  _ChipConfig _chipConfig(String status) {
    switch (status) {
      case 'PENDING':
        return const _ChipConfig(
          color: Colors.amber,
          icon: Icons.schedule,
          label: 'Processing',
        );
      case 'NEEDS_REVIEW':
        return const _ChipConfig(
          color: Colors.orange,
          icon: Icons.edit,
          label: 'Needs Review',
        );
      case 'CONFIRMED':
        return const _ChipConfig(
          color: Colors.green,
          icon: Icons.check_circle,
          label: 'Confirmed',
        );
      case 'REJECTED':
        return const _ChipConfig(
          color: Colors.red,
          icon: Icons.cancel,
          label: 'Rejected',
        );
      default:
        return const _ChipConfig(
          color: Colors.grey,
          icon: Icons.help,
          label: 'Unknown',
        );
    }
  }
}

class _ChipConfig {
  const _ChipConfig({
    required this.color,
    required this.icon,
    required this.label,
  });

  final Color color;
  final IconData icon;
  final String label;
}
