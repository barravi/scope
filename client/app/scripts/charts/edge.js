const React = require('react');
const Spring = require('react-motion').Spring;

const AppActions = require('../actions/app-actions');

const Edge = React.createClass({

  getPath: function(points) {
    const path = 'M' +
      points[0] + ',' +
      points[1] + 'L' +
      points[2] + ',' +
      points[3];

    return path;
  },

  render: function() {
    const className = this.props.highlighted ? 'edge highlighted' : 'edge';
    const path = this.getPath;
    const points = [this.props.points[0].x,
      this.props.points[0].y,
      this.props.points[1].x,
      this.props.points[1].y];

    return (
      <g className={className} onMouseEnter={this.handleMouseEnter} onMouseLeave={this.handleMouseLeave} id={this.props.id}>
        <Spring endValue={points}>
          {function(tweeningPoints) {
            return (
              <path d={path(tweeningPoints)} className="link" />
            );
          }}
        </Spring>
      </g>
    );
  },

  handleMouseEnter: function(ev) {
    AppActions.enterEdge(ev.currentTarget.id);
  },

  handleMouseLeave: function(ev) {
    AppActions.leaveEdge(ev.currentTarget.id);
  }

});

module.exports = Edge;
