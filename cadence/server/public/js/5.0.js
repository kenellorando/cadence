$(document).ready(function() {
    $('#tabs li').on('click', function() {
        var tab = $(this).data('tab');

        $('#tabs li').removeClass('is-active');
        $(this).addClass('is-active');

        $('#tab-content section').removeClass('is-active');
        $('section[data-content="' + tab + '"]').addClass('is-active');
    });
});