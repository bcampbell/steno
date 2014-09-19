  // data from table

  var rows = document.querySelectorAll('#data tbody tr');
  var data = [];
  for (var i=0; i<rows.length; i++){
    var cols = rows[i].querySelectorAll('td');
    data.push({day:new Date(cols[0].childNodes[0].data),
      cnt:parseInt(cols[1].childNodes[0].data)});
  }


// draw the chart
  var barWidth = 20;
  var margin = {top: 20, right: 30, bottom: 30, left: 40},
    width = data.length*barWidth,
    height = 500 - margin.top - margin.bottom;

  var y = d3.scale.linear()
    .domain([0, d3.max(data, function(d){return d.cnt;})])
    .range([height,0]);


  var x = d3.time.scale()
    .domain([
        d3.min(data, function(d){return d.day;}),
        d3.max(data, function(d){return d.day;})
        ])
    .range([barWidth/2,width-(barWidth/2)]);


  var chart = d3.select(".chart")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");


  var bar = chart.selectAll("g")
    .data(data)
    .enter().append("g")
    .attr("transform", function(d, i) { return "translate(" + (i*barWidth) + ",0)"; });

  var yAxis = d3.svg.axis()
    .scale(y)
    .orient("left");

  chart.append("g")
    .attr("class", "y axis")
    .call(yAxis)
    .append("text")
    .attr("transform", "rotate(-90)")
    .attr("y", 6)
    .attr("dy", ".71em")
    .style("text-anchor", "end")
    .text("Matches");

  var xAxis = d3.svg.axis()
    .scale(x)
    .orient("bottom");

  chart.append("g")
      .attr("class", "x axis")
      .attr("transform", "translate(0," + height + ")")
      .call(xAxis);

  bar.append("rect")
    .attr("class", function(d) {
        return d.day.getDay()==0 ? "sunday":""
    })
    .attr("y", function(d) { return y(d.cnt); })
    .attr("width", barWidth-1)
    .attr("height", function(d) { return height - y(d.cnt); });

  bar.append("text")
    .attr("x", barWidth/2)
    .attr("y", function(d) { return y(d.cnt) + 3; })
    .attr("dy", "1em")
    .text(function(d) { return d.cnt==0?"":d.cnt; })

// hide the table of raw data
document.getElementById('data').style.display = 'none';





