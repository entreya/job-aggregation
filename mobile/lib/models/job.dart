class Job {
  final String id;
  final String title;
  final String department;
  final String location;
  final int postedDate; // Unix timestamp
  final String url;

  Job({
    required this.id,
    required this.title,
    required this.department,
    required this.location,
    required this.postedDate,
    required this.url,
  });

  // Factory constructor to create a Job from a Map (Map<String, dynamic> from sqflite)
  factory Job.fromMap(Map<String, dynamic> map) {
    return Job(
      id: map['id'] as String,
      title: map['title'] as String,
      department: map['department'] as String,
      location: map['location'] as String,
      postedDate: map['posted_date'] as int,
      url: map['url'] as String,
    );
  }

  // Returns relative time string (e.g., "2 hours ago")
  String get timeAgo {
    final date = DateTime.fromMillisecondsSinceEpoch(postedDate * 1000);
    final now = DateTime.now();
    final difference = now.difference(date);

    if (difference.inDays > 0) {
      return '${difference.inDays}d ago';
    } else if (difference.inHours > 0) {
      return '${difference.inHours}h ago';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes}m ago';
    } else {
      return 'Just now';
    }
  }
}
