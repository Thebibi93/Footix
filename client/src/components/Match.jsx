function Match({ homeTeam, awayTeam, location, startDate }) {
  const formattedDate = new Date(startDate).toLocaleString("fr-FR", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div className="bg-white rounded-xl p-4 shadow-md border border-gray-200 hover:shadow-lg transition-shadow duration-200">
      <div className="flex justify-between items-center text-xl font-bold mb-3">
        <span className="text-gray-800">{homeTeam}</span>
        <span className="text-sm bg-gray-200 text-gray-700 px-2 py-0.5 rounded-full">vs</span>
        <span className="text-gray-800">{awayTeam}</span>
      </div>
      <div className="flex flex-wrap justify-between items-center gap-2 text-sm text-gray-500 pt-2 border-t border-gray-100">
        <span>📍 {location}</span>
        <span>📅 {formattedDate}</span>
      </div>
    </div>
  );
}

export default Match;